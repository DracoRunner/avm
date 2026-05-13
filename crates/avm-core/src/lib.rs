pub mod config;
pub mod resolver;

pub use config::{load_config, load_with_env, save_config, save_flat_legacy, write_default_config};
pub use config::{ConfigFile, ConfigLoadResult};
pub use resolver::{AliasSource, ResolvedAliasLookup, ResolvedConfig, Resolver};

pub type AliasMap = std::collections::HashMap<String, String>;
pub type ToolMap = std::collections::HashMap<String, String>;
pub type EnvMap = std::collections::HashMap<String, String>;

pub const LOCAL_CONFIG_FILE: &str = ".avm.json";
