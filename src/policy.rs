use std::{collections::HashSet, fs, path::Path};

use serde::Deserialize;

use crate::proto::InvocationContext;

#[derive(Debug, thiserror::Error)]
pub enum PolicyError {
    #[error("failed to read policy: {0}")]
    Read(#[from] std::io::Error),
    #[error("invalid policy: {0}")]
    Parse(#[from] toml::de::Error),
}

pub trait Authorizer: Send + Sync {
    fn authorize(&self, context: &InvocationContext, capability: &str) -> bool;
}

#[derive(Deserialize)]
struct PolicyFile {
    #[serde(default)]
    grants: Vec<Grant>,
}

#[derive(Deserialize)]
struct Grant {
    agent_id: String,
    skill_name: String,
    #[serde(default)]
    skill_version: Option<String>,
    capabilities: HashSet<String>,
}

pub struct StaticAuthorizer {
    grants: Vec<Grant>,
}

impl StaticAuthorizer {
    pub fn load(path: impl AsRef<Path>) -> Result<Self, PolicyError> {
        let source = fs::read_to_string(path)?;
        Self::from_toml(&source)
    }

    pub fn from_toml(source: &str) -> Result<Self, PolicyError> {
        let policy: PolicyFile = toml::from_str(source)?;
        Ok(Self {
            grants: policy.grants,
        })
    }
}

impl Authorizer for StaticAuthorizer {
    fn authorize(&self, context: &InvocationContext, capability: &str) -> bool {
        self.grants.iter().any(|grant| {
            grant.agent_id == context.agent_id
                && grant.skill_name == context.skill_name
                && grant
                    .skill_version
                    .as_ref()
                    .is_none_or(|version| version == &context.skill_version)
                && grant.capabilities.contains(capability)
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn binds_capability_to_agent_skill_and_version() {
        let policy = StaticAuthorizer::from_toml(
            r#"[[grants]]
agent_id = "agent-a"
skill_name = "discovery"
skill_version = "1.0.0"
capabilities = ["scan.passive"]"#,
        )
        .unwrap();
        let mut context = InvocationContext {
            request_id: "req".to_owned(),
            idempotency_key: "idem".to_owned(),
            agent_id: "agent-a".to_owned(),
            skill_name: "discovery".to_owned(),
            skill_version: "1.0.0".to_owned(),
        };
        assert!(policy.authorize(&context, "scan.passive"));
        context.skill_version = "1.0.1".to_owned();
        assert!(!policy.authorize(&context, "scan.passive"));
        assert!(!policy.authorize(&context, "scan.active"));
    }
}
