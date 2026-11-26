//! Agent-scope.
//!
//! Port av `src/api/internal/middleware/agent_scope.go`. Allowlisten är
//! medvetet liten: agenten får arbeta med mål och agentkörningar, men får
//! aldrig besluta om ett godkännande - det är human gate, och den ska
//! alltid göras av en människa. Den får inte heller radera något, eftersom
//! allt som en människa har skapat ska överleva.

/// Svarar på om en agent-nyckel får göra anropet.
pub fn agent_scope_allows(method: &str, path: &str) -> bool {
    let segments: Vec<&str> = path.trim_matches('/').split('/').collect();
    if segments.len() < 3 || segments[0] != "api" || segments[1] != "v1" {
        return false;
    }

    let resource = segments[2];
    let rest = &segments[3..];

    match resource {
        "statuses" => method == "GET",

        // Agenten behöver känna till användarens platser för att kunna
        // föreslå platser på ett mål. Den får inte skapa eller radera
        // platser.
        "places" => method == "GET",

        "goals" => allows_goals(method, rest),

        "agent-runs" => allows_agent_runs(method, rest),

        // Läsa ett godkännande är ok, att besluta om det är det inte.
        "agent-approvals" => method == "GET" && rest.len() == 1,

        _ => false,
    }
}

/// Gäller målen.
fn allows_goals(method: &str, rest: &[&str]) -> bool {
    match rest.len() {
        0 => {
            // Lista och skapa mål - undermål skapas genom Plan, inte genom
            // filsystemet.
            method == "GET" || method == "POST"
        }
        1 => {
            // "relations" är en samling, inte ett mål-id. Dit lägger
            // agenten undermål - ett skrivande anrop som hör ihop med
            // målen, så det tillåts.
            if rest[0] == "relations" {
                return method == "POST";
            }
            method == "GET" || method == "PATCH"
        }
        2 => match rest[1] {
            // Statusen på ett mål.
            "status" => method == "GET" || method == "PATCH",
            // Agenten får sätta taggar och platser på ett mål - de är en
            // del av måldefinitionen. Den får inte radera dem.
            "tags" | "places" => method == "PUT",
            _ => false,
        },
        _ => false,
    }
}

/// Gäller agentkörningarna och deras händelser och godkännanden.
fn allows_agent_runs(method: &str, rest: &[&str]) -> bool {
    match rest.len() {
        0 => method == "GET" || method == "POST",
        1 => method == "GET" || method == "PATCH",
        2 => match rest[1] {
            "events" => method == "GET" || method == "POST",
            "approvals" => method == "GET" || method == "POST",
            // Agenten rapporterar sin användning. Den får inte radera
            // eller ändra den i efterhand.
            "usage" => method == "GET" || method == "POST",
            _ => false,
        },
        _ => false,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn laser_statusar_och_platser() {
        assert!(agent_scope_allows("GET", "/api/v1/statuses"));
        assert!(!agent_scope_allows("POST", "/api/v1/statuses"));
        assert!(agent_scope_allows("GET", "/api/v1/places"));
        assert!(!agent_scope_allows("DELETE", "/api/v1/places/1"));
    }

    #[test]
    fn laser_mal_utan_radering() {
        assert!(agent_scope_allows("GET", "/api/v1/goals"));
        assert!(agent_scope_allows("POST", "/api/v1/goals"));
        assert!(agent_scope_allows("GET", "/api/v1/goals/12"));
        assert!(agent_scope_allows("PATCH", "/api/v1/goals/12"));
        assert!(!agent_scope_allows("DELETE", "/api/v1/goals/12"));
        assert!(agent_scope_allows("GET", "/api/v1/goals/12/status"));
        assert!(agent_scope_allows("PATCH", "/api/v1/goals/12/status"));
        assert!(agent_scope_allows("PUT", "/api/v1/goals/12/tags"));
        assert!(agent_scope_allows("PUT", "/api/v1/goals/12/places"));
        assert!(!agent_scope_allows("DELETE", "/api/v1/goals/12/tags/ute"));
    }

    #[test]
    fn laser_undermal_genom_relationer() {
        assert!(agent_scope_allows("POST", "/api/v1/goals/relations"));
        assert!(!agent_scope_allows("DELETE", "/api/v1/goals/relations"));
    }

    #[test]
    fn laser_agentkorningar() {
        assert!(agent_scope_allows("GET", "/api/v1/agent-runs"));
        assert!(agent_scope_allows("POST", "/api/v1/agent-runs"));
        assert!(agent_scope_allows("GET", "/api/v1/agent-runs/7"));
        assert!(agent_scope_allows("PATCH", "/api/v1/agent-runs/7"));
        assert!(agent_scope_allows("POST", "/api/v1/agent-runs/7/events"));
        assert!(agent_scope_allows("POST", "/api/v1/agent-runs/7/approvals"));
        assert!(agent_scope_allows("POST", "/api/v1/agent-runs/7/usage"));
        assert!(!agent_scope_allows("DELETE", "/api/v1/agent-runs/7"));
    }

    #[test]
    fn laser_inte_besluta_om_godkannande() {
        // Bara ett enskilt godkännande får läsas - listan är förbjuden,
        // likadeles att besluta om ett godkännande.
        assert!(!agent_scope_allows("GET", "/api/v1/agent-approvals"));
        assert!(agent_scope_allows("GET", "/api/v1/agent-approvals/3"));
        assert!(!agent_scope_allows(
            "POST",
            "/api/v1/agent-approvals/3/decision"
        ));
    }

    #[test]
    fn allt_annat_ar_forbudet() {
        assert!(!agent_scope_allows("GET", "/api/v1/calendars"));
        assert!(!agent_scope_allows("GET", "/api/v1/calendar-events"));
        assert!(!agent_scope_allows("GET", "/api/v1/tags"));
        assert!(!agent_scope_allows("GET", "/api/v1/api-keys"));
        assert!(!agent_scope_allows("GET", "/health"));
    }
}
