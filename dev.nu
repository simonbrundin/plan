
# -----------------------------------------------
# Aktivera Mise-paket
# -----------------------------------------------

def activate-mise [] {
  let hook_output = (^mise hook-env -s nu | lines | split column "," | rename op name value)
  for line in $hook_output {
    if ($line.op == "set" and $line.name == "PATH") {
      $env.PATH = ($line.value | split row (char esep) | append $env.PATH)
    }
  }
}

activate-mise

# -----------------------------------------------
# Funktioner
# -----------------------------------------------

def get-teleport-username [] {
    let status = (tsh status | lines)
    let username_line = ($status | where $it =~ "Logged in as:" | first)
    let username = ($username_line | split row ":" | last | str trim)
    $username
}

# -----------------------------------------------
# Variables
# -----------------------------------------------

let teleport_user = (get-teleport-username)
let namespace = $"plan-dev-($teleport_user)" 
let link = $"https://($namespace).simonbrundin.com"

# -----------------------------------------------
# Skicka notis till telefon
# -----------------------------------------------

curl -d $link ntfy.sh/simonbrundin-dev-notification

# -----------------------------------------------
# Kör Tilt
# -----------------------------------------------

cd .devcontainer/

# Kontrollera om kubectl fungerar, annars logga in via teleport
try {
  kubectl cluster-info | ignore
} catch {
  simon kubernetes login teleport
}

# Avsluta befintliga Tilt-processer och tjänster på port 10350
try {
  # Avsluta alla körande Tilt-processer
  let tilt_pids = (pgrep tilt | lines | where $it != "")
  if ($tilt_pids | length) > 0 {
    print $"Avslutar befintliga Tilt-processer: ($tilt_pids)"
    pkill -9 tilt
  }
} catch {
  # Inga Tilt-processer hittades
}

try {
  # Avsluta allt som lyssnar på port 10350
  fuser -k 10350/tcp
} catch {
  # Ingen process på porten
}

tilt up --namespace $namespace


