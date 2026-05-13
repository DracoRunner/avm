use anyhow::{anyhow, Context, Result};
use avm_plugin_api::{AliasDetail, AliasValue, ExportResponse, Manifest, ResolvedAlias};
use std::collections::HashMap;
use std::fs;
use std::io::Read;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};
use std::time::{Duration, Instant};
use wait_timeout::ChildExt;

const PLUGIN_TIMEOUT_MS: u64 = 500;
const GLOBAL_TIMEOUT_MS: u64 = 1000;

#[derive(Debug)]
pub struct PluginManager {
    plugin_dir: PathBuf,
}

impl PluginManager {
    pub fn new(plugin_dir: Option<PathBuf>) -> Result<Self> {
        let dir = plugin_dir.unwrap_or_else(default_plugin_dir);
        fs::create_dir_all(&dir).context("create plugin directory")?;
        Ok(Self { plugin_dir: dir })
    }

    pub fn plugin_dir(&self) -> PathBuf {
        self.plugin_dir.clone()
    }

    pub fn list_aliases(&self, cwd: &Path) -> Result<HashMap<String, ResolvedAlias>> {
        if !self.plugin_dir.exists() {
            return Ok(HashMap::new());
        }

        let mut entries: Vec<_> = fs::read_dir(&self.plugin_dir)
            .context("unable to read plugin directory")?
            .filter_map(Result::ok)
            .filter(|entry| {
                let ft = entry.file_type().ok();
                ft.map(|ty| ty.is_dir()).unwrap_or(false)
            })
            .collect();

        entries.sort_by_key(|entry| entry.file_name().to_string_lossy().to_ascii_lowercase());

        let mut result = HashMap::new();
        let start = Instant::now();
        let global_timeout = Duration::from_millis(GLOBAL_TIMEOUT_MS);

        for entry in entries {
            if start.elapsed() > global_timeout {
                break;
            }

            let plugin_name = entry.file_name().to_string_lossy().to_string();
            let plugin_path = entry.path();
            match load_plugin_aliases(&plugin_path, cwd) {
                Ok(aliases) => {
                    for (key, alias) in aliases {
                        // First plugin wins while preserving directory sort order.
                        result.entry(key).or_insert(alias);
                    }
                }
                Err(err) if std::env::var("AVM_DEBUG").ok().as_deref() == Some("1") => {
                    eprintln!("[avm] plugin {plugin_name} skipped: {err}");
                }
                Err(_) => {}
            }
        }

        Ok(result)
    }

    pub fn list_plugins(&self) -> Result<HashMap<String, Manifest>> {
        if !self.plugin_dir.exists() {
            return Ok(HashMap::new());
        }

        let mut plugins = HashMap::new();
        for entry in fs::read_dir(&self.plugin_dir).context("failed reading plugin dir")? {
            let entry = match entry {
                Ok(e) => e,
                Err(_) => continue,
            };

            let ty = match entry.file_type() {
                Ok(ft) => ft,
                Err(_) => continue,
            };
            if !ty.is_dir() {
                continue;
            }

            let name = entry.file_name().to_string_lossy().to_string();
            if let Ok(manifest) = read_manifest_path(&entry.path()) {
                plugins.insert(name, manifest);
            }
        }

        Ok(plugins)
    }

    pub fn read_manifest(&self, name: &str) -> Result<Manifest> {
        let plugin_path = self.plugin_dir.join(name);
        read_manifest_path(&plugin_path)
    }

    pub fn install_plugin(&self, source: &str) -> Result<()> {
        let is_remote = is_git_url(source);
        let target = if is_remote {
            let plugin_name = derive_remote_plugin_name(source)?;
            self.plugin_dir.join(plugin_name)
        } else {
            let source_metadata = fs::symlink_metadata(source).context("invalid plugin source path")?;
            if source_metadata.file_type().is_symlink() {
                return Err(anyhow!("plugin source must not be a symlink"));
            }

            let source_dir = fs::canonicalize(source).context("invalid plugin source path")?;
            validate_plugin_source_permissions(&source_dir)?;
            let source_name = source_dir
                .file_name()
                .and_then(|name| name.to_str())
                .ok_or_else(|| anyhow!("invalid plugin source"))?;
            self.plugin_dir.join(source_name)
        };

        if target.exists() {
            return Err(anyhow!(
                "plugin already installed; use `avm plugin update <name>` first",
            ));
        }

        let install_result = if is_remote {
            let status = Command::new("git")
                .arg("clone")
                .arg(source)
                .arg(&target)
                .env_clear()
                .env("PATH", default_plugin_path_env())
                .status()
                .context("git clone failed")?;
            if !status.success() {
                Err(anyhow!("git clone failed with status {}", status))
            } else {
                Ok(())
            }
        } else {
            let source = fs::canonicalize(source).context("invalid plugin source path")?;
            copy_dir_recursive(&source, &target).context("plugin copy failed")
        };

        if let Err(err) = install_result {
            let _ = fs::remove_dir_all(&target);
            return Err(err);
        }

        let manifest_path = target.join("plugin.json");
        if !manifest_path.exists() {
            let _ = fs::remove_dir_all(&target);
            return Err(anyhow!("invalid plugin: missing plugin.json"));
        }

        if !target.join("bin").join("export-aliases").exists() {
            let _ = fs::remove_dir_all(&target);
            return Err(anyhow!("invalid plugin: missing bin/export-aliases"));
        }

        Ok(())
    }

    pub fn remove_plugin(&self, name: &str) -> Result<()> {
        let target = self.plugin_dir.join(name);
        if target.exists() {
            fs::remove_dir_all(target).context("remove plugin")?;
        }
        Ok(())
    }

    pub fn update_plugin(&self, name: &str) -> Result<()> {
        let target = self.plugin_dir.join(name);
        if !target.exists() {
            return Err(anyhow!("plugin '{}' not found", name));
        }

        if !target.join(".git").exists() {
            return Ok(());
        }

        let status = Command::new("git")
            .arg("-C")
            .arg(&target)
            .arg("pull")
            .env_clear()
            .env("PATH", default_plugin_path_env())
            .status()
            .context("git pull failed")?;
        if !status.success() {
            return Err(anyhow!("plugin update failed with status {}", status));
        }

        Ok(())
    }
}

fn default_plugin_dir() -> PathBuf {
    if let Ok(home) = std::env::var("AVM_PLUGIN_DIR") {
        return PathBuf::from(home);
    }

    if let Ok(home) = std::env::var("HOME") {
        return PathBuf::from(home).join(".avm").join("plugins");
    }

    PathBuf::from(".").join(".avm").join("plugins")
}

fn read_manifest_path(path: &Path) -> Result<Manifest> {
    let raw = fs::read_to_string(path.join("plugin.json")).context("unable to read plugin manifest")?;
    let manifest: Manifest = serde_json::from_str(&raw).context("invalid plugin manifest")?;
    Ok(manifest)
}

fn validate_plugin_source_permissions(path: &Path) -> Result<()> {
    let meta = fs::metadata(path).context("failed to read plugin source metadata")?;
    if !meta.is_dir() {
        return Err(anyhow!("plugin source must be a directory"));
    }

    #[cfg(unix)]
    {
        use std::os::unix::fs::MetadataExt;
        use std::os::unix::fs::PermissionsExt;
        if meta.uid() == 0 {
            return Err(anyhow!("plugin sources owned by root are not allowed"));
        }
        let mode = meta.permissions().mode();
        if mode & 0o002 != 0 {
            return Err(anyhow!("plugin source must not be world-writable"));
        }
    }

    Ok(())
}

fn derive_remote_plugin_name(source: &str) -> Result<String> {
    let trimmed = source.trim();
    if trimmed.is_empty() {
        return Err(anyhow!("plugin source is empty"));
    }

    let mut candidate = trimmed;
    if candidate.starts_with("git@") {
        candidate = candidate.split_once(':').map(|(_, tail)| tail).unwrap_or(candidate);
    }

    candidate = candidate.split(['#', '?']).next().unwrap_or(candidate).trim_end_matches('/');
    let name = candidate
        .split('/')
        .filter(|part| !part.is_empty())
        .last()
        .ok_or_else(|| anyhow!("unable to derive plugin name"))?;

    let name = name.strip_suffix(".git").unwrap_or(name);
    if name.is_empty() || name.contains('\0') || name.contains('/') || name.contains('\\') {
        return Err(anyhow!("invalid plugin name derived from source"));
    }
    if name.chars().any(|ch| ch.is_control()) {
        return Err(anyhow!("invalid plugin name derived from source"));
    }

    Ok(name.to_string())
}

fn copy_dir_recursive(source: &Path, destination: &Path) -> Result<()> {
    fs::create_dir_all(destination).context("failed to create plugin destination")?;
    for entry in fs::read_dir(source).context("failed to read plugin source")? {
        let entry = entry?;
        let file_type = entry.file_type()?;
        let src = entry.path();
        let dst = destination.join(entry.file_name());

        if file_type.is_symlink() {
            return Err(anyhow!(
                "plugin source contains symlink entries; remove symlinks before installing"
            ));
        }

        if file_type.is_dir() {
            copy_dir_recursive(&src, &dst)?;
            continue;
        }

        if file_type.is_file() {
            fs::copy(&src, &dst).with_context(|| format!("failed to copy {:?}", src))?;
            continue;
        }

        return Err(anyhow!("unsupported plugin source entry type: {:?}", src));
    }
    Ok(())
}

fn sandbox_plugin_command(cmd: &mut Command, plugin_path: &Path) {
    cmd.env_clear();
    cmd.current_dir(plugin_path);
    cmd.env("AVM_PLUGIN_DIR", plugin_path);
    cmd.env("PATH", default_plugin_path_env());
}

#[cfg(unix)]
fn default_plugin_path_env() -> &'static str {
    "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
}

#[cfg(not(unix))]
fn default_plugin_path_env() -> &'static str {
    r#"C:\Windows\system32;C:\Windows;C:\Windows\System32\Wbem;C:\Windows\System32\WindowsPowerShell\v1.0\"#
}

fn run_with_timeout(mut cmd: Command, timeout_ms: u64) -> Result<String> {
    let mut child = cmd
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .context("failed to start plugin command")?;

    let timeout = Duration::from_millis(timeout_ms);
    let status = child
        .wait_timeout(timeout)
        .context("failed while waiting for plugin command")?;
    let status = match status {
        Some(status) => status,
        None => {
            let _ = child.kill();
            let _ = child.wait();
            return Err(anyhow!("plugin command timed out"));
        }
    };

    let mut out = String::new();
    let mut err = String::new();
    if let Some(mut stdout) = child.stdout.take() {
        let _ = stdout.read_to_string(&mut out);
    }
    if let Some(mut stderr) = child.stderr.take() {
        let _ = stderr.read_to_string(&mut err);
    }

    if !status.success() {
        if !err.trim().is_empty() {
            return Err(anyhow!("plugin command failed: {err}"));
        }
        return Err(anyhow!("plugin command failed with exit code"));
    }

    Ok(out)
}

fn normalize_section(manifest: &Manifest) -> String {
    manifest
        .section_label
        .clone()
        .unwrap_or_else(|| manifest.name.clone())
}

fn load_plugin_aliases(plugin_path: &Path, cwd: &Path) -> Result<HashMap<String, ResolvedAlias>> {
    let manifest = read_manifest_path(plugin_path)?;

    let wasm_hook = plugin_path.join("bin").join("export-aliases.wasm");
    let bin_hook = plugin_path.join("bin").join("export-aliases");

    if !bin_hook.exists() && !wasm_hook.exists() {
        return Err(anyhow!("missing export-aliases"));
    }

    let hook_output = if bin_hook.exists() {
        let health = plugin_path.join("bin").join("health-check");
        if health.exists() {
            let mut cmd = Command::new(health);
            cmd.arg("--dir").arg(cwd);
            sandbox_plugin_command(&mut cmd, plugin_path);
            if run_with_timeout(cmd, PLUGIN_TIMEOUT_MS).is_err() {
                return Err(anyhow!("plugin health-check failed"));
            }
        }

        let mut cmd = Command::new(bin_hook);
        cmd.arg("--dir").arg(cwd);
        sandbox_plugin_command(&mut cmd, plugin_path);
        run_with_timeout(cmd, PLUGIN_TIMEOUT_MS)?
    } else {
        return Err(anyhow!(
            "wasm plugin execution is not enabled in this baseline; please keep node scripts in the merged provider"
        ));
    };

    if hook_output.trim().is_empty() {
        return Ok(HashMap::new());
    }

    let response: ExportResponse = serde_json::from_str(&hook_output).context("invalid plugin response")?;
    let mut aliases = HashMap::new();
    let section = normalize_section(&manifest);

    for (key, value) in response.aliases {
        let mapped = match value {
            AliasValue::Simple(command) => AliasDetail {
                command,
                description: None,
                source: Some("plugin".to_string()),
            },
            AliasValue::Detailed(detail) => detail,
        };

        if mapped.command.trim().is_empty() {
            continue;
        }

        aliases.insert(
            key,
            ResolvedAlias {
                command: mapped.command,
                description: mapped.description,
                plugin_name: manifest.name.clone(),
                section_name: section.clone(),
                source: mapped.source,
            },
        );
    }

    Ok(aliases)
}

fn is_git_url(source: &str) -> bool {
    source.starts_with("https://")
        || source.starts_with("http://")
        || source.starts_with("git@")
        || source.starts_with("git://")
        || source.starts_with("ssh://")
}
