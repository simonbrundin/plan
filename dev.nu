#!/usr/bin/env nu

# -----------------------------------------------
# Aktivera Mise-paket
# -----------------------------------------------

def activate-mise [] {
  let mise_path = (^mise env | parse -r 'export PATH=\'(.+)\'' | get capture0.0 | split row (char esep))
  $env.PATH = ($mise_path | append $env.PATH)
}

activate-mise



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

let local_ip = (get-local-ip)
let namespace = "plan-dev" 
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

# Kontrollera om kubectl fungerar
let kubectl_check = (kubectl cluster-info | complete)

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
# Välj Tilt mode med fzf
# -----------------------------------------------

print "\n🚀 Välj utvecklingsmiljö:\n"

let modes = [
  "local - Docker Compose (standard)",
  "kubernetes - Full cluster miljö"
]

let selection = ($modes | fzf --prompt="Mode: " --height=40% --reverse | str trim)

let mode = if ($selection | str contains "kubernetes") {
  "kubernetes"
} else {
  "local"
}

print $"\n✓ Startar Tilt i ($mode) mode...\n"

print $"\n✓ Startar Tilt i ($mode) mode...\n"

if $mode == "kubernetes" {
  tilt up --namespace $namespace -- mode=kubernetes
} else {
  tilt up --namespace $namespace -- mode=local
}


