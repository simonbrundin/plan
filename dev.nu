
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

def get-local-ip [] {
    try {
        # Hämta den primära lokala IP-adressen som används för internet-trafik
        let ip_output = (ip route get 1.1.1.1 | complete)
        if $ip_output.exit_code == 0 {
            let ip = ($ip_output.stdout | str trim | parse -r 'src\s+(\S+)' | get capture0.0)
            $ip
        } else {
            # Fallback: använd hostname -I och ta första IP-adressen
            let ips = (hostname -I | str trim | split row ' ')
            $ips.0
        }
    } catch {
        # Om allt annat misslyckas, returnera localhost
        "127.0.0.1"
    }
}

# -----------------------------------------------
# Variables
# -----------------------------------------------

let teleport_user = (get-teleport-username)
let local_ip = (get-local-ip)
let namespace = $"plan-dev-($teleport_user)" 
# let link = $"https://($namespace).simonbrundin.com"
let link = $"http://($local_ip):3000"

# -----------------------------------------------
# Skicka notis till telefon
# -----------------------------------------------

curl -d $link ntfy.sh/simonbrundin-dev-notification

# -----------------------------------------------
# Kör Tilt
# -----------------------------------------------

cd .devcontainer/

# Kontrollera om kubectl fungerar, annars logga in via teleport
let kubectl_check = (kubectl cluster-info | complete)
if $kubectl_check.exit_code != 0 {
  print "Kan inte nå Kubernetes, loggar in via Teleport..."
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

# -----------------------------------------------
# Välj Tilt mode
# -----------------------------------------------

print "\n🚀 Välj utvecklingsmiljö:\n"

let modes = [
  "local - Docker Compose (standard)",
  "kubernetes - Full cluster miljö"
]

let selection = ($modes | str join "\n" | fzf --prompt="Mode: " --height=40% --reverse | str trim)

let mode = if ($selection | str contains "kubernetes") {
  "kubernetes"
} else {
  "local"
}

print $"\n✓ Startar Tilt i ($mode) mode...\n"

if $mode == "kubernetes" {
  tilt up --namespace $namespace -- mode=kubernetes
} else {
  tilt up --namespace $namespace -- mode=local
}


