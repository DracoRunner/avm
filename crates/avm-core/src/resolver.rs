use std::collections::HashMap;
use std::path::PathBuf;

use avm_plugin_api::ResolvedAlias;

#[derive(Debug, Clone)]
pub enum AliasSource {
    Local,
    Global,
    Plugin,
}

#[derive(Debug, Clone)]
pub struct ResolvedAliasLookup {
    pub command: String,
    pub source: AliasSource,
    pub plugin_name: Option<String>,
}

#[derive(Debug, Clone)]
pub struct ResolvedConfig {
    pub local_aliases: HashMap<String, String>,
    pub global_aliases: HashMap<String, String>,
    pub local_env: HashMap<String, String>,
    pub global_env: HashMap<String, String>,
    pub local_tools: HashMap<String, String>,
    pub global_tools: HashMap<String, String>,
    pub plugin_aliases: HashMap<String, ResolvedAlias>,
}

impl ResolvedConfig {
    pub fn resolve_alias(&self, key: &str, _cfg: &ResolvedConfig) -> Option<ResolvedAliasLookup> {
        if let Some(value) = self.local_aliases.get(key) {
            return Some(ResolvedAliasLookup {
                command: value.clone(),
                source: AliasSource::Local,
                plugin_name: None,
            });
        }

        if let Some(value) = self.global_aliases.get(key) {
            return Some(ResolvedAliasLookup {
                command: value.clone(),
                source: AliasSource::Global,
                plugin_name: None,
            });
        }

        self.plugin_aliases.get(key).map(|a| ResolvedAliasLookup {
            command: a.command.clone(),
            source: AliasSource::Plugin,
            plugin_name: Some(a.plugin_name.clone()),
        })
    }

    pub fn resolve_tool(&self, key: &str, _cfg: &ResolvedConfig) -> Option<(String, AliasSource)> {
        if let Some(version) = self.local_tools.get(key) {
            return Some((version.clone(), AliasSource::Local));
        }
        if let Some(version) = self.global_tools.get(key) {
            return Some((version.clone(), AliasSource::Global));
        }
        None
    }

    pub fn resolve_tools_with_source(
        &self,
        _cfg: &ResolvedConfig,
    ) -> HashMap<String, (String, AliasSource)> {
        let mut merged: HashMap<String, (String, AliasSource)> = HashMap::new();
        for (tool, version) in &self.global_tools {
            merged.insert(tool.clone(), (version.clone(), AliasSource::Global));
        }
        for (tool, version) in &self.local_tools {
            merged.insert(tool.clone(), (version.clone(), AliasSource::Local));
        }
        merged
    }
}

#[derive(Debug, Clone)]
pub struct Resolver {
    cwd: PathBuf,
    home: PathBuf,
    config_file: String,
}

impl Resolver {
    pub fn new(cwd: PathBuf, home: PathBuf) -> Self {
        Self {
            cwd,
            home,
            config_file: ".avm.json".to_string(),
        }
    }

    pub fn load(&self, plugin_aliases: HashMap<String, ResolvedAlias>) -> anyhow::Result<ResolvedConfig> {
        let local = crate::config::load_with_env(&self.cwd, &self.config_file)?;
        let global = crate::config::load_with_env(&self.home, &self.config_file)?;

        Ok(ResolvedConfig {
            local_aliases: local.aliases,
            global_aliases: global.aliases,
            local_env: local.env,
            global_env: global.env,
            local_tools: local.tools,
            global_tools: global.tools,
            plugin_aliases,
        })
    }

    pub fn resolve_alias(&self, key: &str, cfg: &ResolvedConfig) -> Option<ResolvedAliasLookup> {
        if let Some(value) = cfg.local_aliases.get(key) {
            return Some(ResolvedAliasLookup {
                command: value.clone(),
                source: AliasSource::Local,
                plugin_name: None,
            });
        }

        if let Some(value) = cfg.global_aliases.get(key) {
            return Some(ResolvedAliasLookup {
                command: value.clone(),
                source: AliasSource::Global,
                plugin_name: None,
            });
        }

        cfg.plugin_aliases.get(key).map(|a| ResolvedAliasLookup {
            command: a.command.clone(),
            source: AliasSource::Plugin,
            plugin_name: Some(a.plugin_name.clone()),
        })
    }

    pub fn resolve_env(&self, cfg: &ResolvedConfig) -> HashMap<String, String> {
        let mut merged = HashMap::new();
        for (k, v) in &cfg.global_env {
            merged.insert(k.clone(), v.clone());
        }
        for (k, v) in &cfg.local_env {
            merged.insert(k.clone(), v.clone());
        }
        merged
    }

    pub fn resolve_tools_with_source(&self, cfg: &ResolvedConfig) -> HashMap<String, (String, AliasSource)> {
        let mut merged: HashMap<String, (String, AliasSource)> = HashMap::new();
        for (tool, version) in &cfg.global_tools {
            merged.insert(tool.clone(), (version.clone(), AliasSource::Global));
        }
        for (tool, version) in &cfg.local_tools {
            merged.insert(tool.clone(), (version.clone(), AliasSource::Local));
        }
        merged
    }

    pub fn resolve_tool(&self, key: &str, cfg: &ResolvedConfig) -> Option<(String, AliasSource)> {
        if let Some(version) = cfg.local_tools.get(key) {
            return Some((version.clone(), AliasSource::Local));
        }
        if let Some(version) = cfg.global_tools.get(key) {
            return Some((version.clone(), AliasSource::Global));
        }
        None
    }

    pub fn suggest_aliases(&self, query: &str, cfg: &ResolvedConfig) -> Vec<String> {
        let mut suggestions = Vec::new();
        let mut candidates = HashMap::new();

        for key in cfg.local_aliases.keys() {
            candidates.insert(key.clone(), true);
        }
        for key in cfg.global_aliases.keys() {
            candidates.insert(key.clone(), true);
        }
        for key in cfg.plugin_aliases.keys() {
            candidates.insert(key.clone(), true);
        }

        let normalized_query = normalize_for_comparison(query);
        for key in candidates.keys() {
            if levenshtein_distance(query, key) <= 2 {
                suggestions.push(key.clone());
                continue;
            }

            let normalized_key = normalize_for_comparison(key);
            if normalized_query == normalized_key && !normalized_query.is_empty() {
                suggestions.push(key.clone());
                continue;
            }

            if (normalized_key.contains(&normalized_query) || normalized_query.contains(&normalized_key))
                && (normalized_key.len() > 3 || normalized_query.len() > 3)
            {
                suggestions.push(key.clone());
            }
        }

        suggestions.sort_unstable();
        suggestions
    }
}

fn normalize_for_comparison(s: &str) -> String {
    let mut parts: Vec<&str> = s
        .split(|c| matches!(c, '-' | ':' | '_' | '.'))
        .filter(|p| !p.is_empty())
        .collect();
    parts.sort_unstable();
    parts.join("-")
}

fn levenshtein_distance(s: &str, t: &str) -> usize {
    let m = s.len();
    let n = t.len();
    let mut dp = vec![vec![0usize; n + 1]; m + 1];

    for i in 0..=m {
        dp[i][0] = i;
    }
    for j in 0..=n {
        dp[0][j] = j;
    }

    let s_bytes = s.as_bytes();
    let t_bytes = t.as_bytes();

    for i in 1..=m {
        for j in 1..=n {
            let cost = if s_bytes[i - 1] == t_bytes[j - 1] { 0 } else { 1 };
            dp[i][j] = std::cmp::min(
                std::cmp::min(dp[i - 1][j] + 1, dp[i][j - 1] + 1),
                dp[i - 1][j - 1] + cost,
            );
        }
    }

    dp[m][n]
}
