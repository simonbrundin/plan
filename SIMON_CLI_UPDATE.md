# Simon CLI Update Required

The "simon dev" command in `/Users/simon/repos/simon-cli/scripts/coding.nu` needs to be updated with better validation and error handling.

## Current Implementation
```nushell
def "main dev" [] {
  print $"(ansi blue) Letar efter .devcontainer.json...(ansi reset)"
  let files = ls -a | where name == ".devcontainer.json"
  if ($files | length) > 0 {
print $"(ansi green) .devcontainer.json hittades!(ansi reset)"
  } else {
print $"(ansi red) .devcontainer.json hittades inte!(ansi reset)"
  }

  let devpodCommand = "devpod up . --provider kubernetes --debug"
print $"(ansi blue)Kör `($devpodCommand)`(ansi reset)"
  nu -c $"SHELL=/bin/bash ($devpodCommand)"
}
```

## Updated Implementation
Replace the above with:

```nushell
def "main dev" [] {
  # Check for required files
  print $"(ansi blue)Checking development environment setup...(ansi reset)"

  let devcontainerExists = (ls -a | where name == ".devcontainer.json" | length) > 0
  let devpodConfigExists = (ls -a | where name == ".devpod.yaml" | length) > 0

  if not $devcontainerExists {
    print $"(ansi red).devcontainer.json not found!(ansi reset)"
    return
  }

  if not $devpodConfigExists {
    print $"(ansi red).devpod.yaml not found!(ansi reset)"
    return
  }

  print $"(ansi green)Configuration files found ✓(ansi reset)"

  # Check if devpod is installed
  let devpodInstalled = (try { devpod version | complete | get exit_code } catch { 1 }) == 0
  if not $devpodInstalled {
    print $"(ansi red)DevPod is not installed. Please install DevPod first.(ansi reset)"
    return
  }

  # Check Kubernetes connectivity
  let kubectlWorks = (try { kubectl cluster-info | complete | get exit_code } catch { 1 }) == 0
  if not $kubectlWorks {
    print $"(ansi red)Cannot connect to Kubernetes cluster. Please check your kubeconfig.(ansi reset)"
    return
  }

  print $"(ansi green)Kubernetes connection verified ✓(ansi reset)"

  let devpodCommand = "devpod up . --provider kubernetes --debug"
  print $"(ansi blue)Starting DevPod workspace with Kubernetes provider...(ansi reset)"
  print $"(ansi blue)Command: ($devpodCommand)(ansi reset)"

  nu -c $"SHELL=/bin/bash ($devpodCommand)"
}
```

## What This Update Does
1. Checks for both `.devcontainer.json` and `.devpod.yaml` files
2. Validates that DevPod is installed
3. Verifies Kubernetes connectivity before attempting to start
4. Provides better error messages and user feedback
5. Ensures all prerequisites are met before running the DevPod command