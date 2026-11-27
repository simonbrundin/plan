//! Agentprofilernas policy.
//!
//! Go: `internal/profile`. Profilen är inte ett stycke text i prompten -
//! den är tekniska gränser som gäller oavsett vad modellen säger.

/// Namnen på profilerna.
pub const CODING: &str = "coding";
/// Profilen för agentarbete som inte rör kod.
pub const LIFE: &str = "life";

/// Vad en profil får göra.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Policy {
    /// Profilens namn.
    pub name: String,
    /// Om profilen får ändra filer i repot.
    pub may_change_code: bool,
    /// Om jobbet behöver klona repot.
    pub needs_repo: bool,
    /// Om körningen alltid kräver ett mänskligt beslut innan något skrivs.
    pub needs_human_gate: bool,
    /// CPU-gränsen för jobbet.
    pub cpu_limit: String,
    /// Minnesgränsen för jobbet.
    pub memory_limit: String,
    /// Hur många rundor modellen får köra.
    pub max_turns: usize,
}

impl Policy {
    /// Policyn för kodändringar.
    ///
    /// Go: `CodingPolicy`.
    pub fn coding() -> Self {
        Self {
            name: CODING.to_string(),
            may_change_code: true,
            needs_repo: true,
            needs_human_gate: true,
            cpu_limit: "1000m".to_string(),
            memory_limit: "2Gi".to_string(),
            max_turns: 20,
        }
    }

    /// Policyn för agentarbete utan kodändringar.
    ///
    /// Go: `LifePolicy`.
    pub fn life() -> Self {
        Self {
            name: LIFE.to_string(),
            may_change_code: false,
            needs_repo: false,
            needs_human_gate: true,
            cpu_limit: "500m".to_string(),
            memory_limit: "1Gi".to_string(),
            max_turns: 12,
        }
    }
}

/// Policyn för en profil.
///
/// En okänd profil ger coding - den är den strängaste, och en okänd
/// profil ska inte ge mer rättigheter än vad som är avsett.
///
/// Go: `PolicyFor`.
pub fn policy_for(name: &str) -> Result<Policy, String> {
    match name.trim().to_lowercase().as_str() {
        "" | CODING => Ok(Policy::coding()),
        LIFE => Ok(Policy::life()),
        annat => Err(format!("unknown agent profile {annat:?}")),
    }
}

/// De profiler som stöds.
///
/// Go: `Names`.
pub fn names() -> Vec<&'static str> {
    vec![CODING, LIFE]
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tom_profil_giver_coding() {
        assert_eq!(policy_for("").unwrap(), Policy::coding());
        assert_eq!(policy_for("coding").unwrap(), Policy::coding());
        assert_eq!(policy_for(" CODING ").unwrap(), Policy::coding());
    }

    #[test]
    fn life_profilen_far_inte_rora_kod() {
        let policy = policy_for("life").unwrap();

        assert!(!policy.may_change_code);
        assert!(!policy.needs_repo);
        assert_eq!(policy.max_turns, 12);
    }

    #[test]
    fn okand_profil_ger_fel() {
        assert_eq!(
            policy_for("pengar"),
            Err("unknown agent profile \"pengar\"".to_string())
        );
    }
}
