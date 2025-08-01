#!/usr/bin/env python3

# Starts a vagrant and executed the passed command
# Requirements:
# * WORKDIR - environment variable pointing to the root of the project
# * arg[1] - vagrant configuration name
# * arg[2] - command to execute


import subprocess
import os
import sys
import atexit

# Get project directory from env var
PROJECT_DIR = os.environ.get("WORKDIR")
if not PROJECT_DIR:
    print("Error: WORKDIR environment variable is not set.")
    sys.exit(1)

LOG_DIR = os.path.join(PROJECT_DIR, "dist", "logs")
os.makedirs(LOG_DIR, exist_ok=True)

def run(cmd, cwd=PROJECT_DIR, check=True):
    print(f"Running: {cmd}")
    result = subprocess.run(cmd, shell=True, cwd=cwd, stdout=sys.stdout, stderr=sys.stderr)
    if check and result.returncode != 0:
        raise RuntimeError(f"Command failed: {cmd}")
    
def run_and_capture(cmd, output_file, cwd=PROJECT_DIR):
    print(f"Capturing: {cmd} → {output_file}")
    with open(output_file, "w") as f:
        subprocess.run(cmd, shell=True, cwd=cwd, stdout=f, stderr=subprocess.STDOUT)

def cleanup(machine_name: str):
    print("[INFO] Cleaning up Vagrant VM...")
    try:
        # Capture NordVPN daemon logs
        log_cmd = f'vagrant ssh {machine_name} -c "journalctl -u snap.nordvpn.nordvpnd.service"'
        run_and_capture(log_cmd, os.path.join(LOG_DIR, "daemon.log"))
        
        # Stop 
        run(f"vagrant halt {machine_name}", check=True)
    except Exception as e:
        print(f"[WARN] Vagrant halt failed: {e}")
    finally:
        try:
            run(f"vagrant destroy {machine_name} --force", check=True)
        except Exception as e:
            print(f"[ERR] Vagrant force destroy failed: {e}")    

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: run-vagrant.py <config-name-or-box> <command-to-execute>")
        sys.exit(1)

    machine_name = sys.argv[1]  # e.g., 'default', 'fedora', 'ubuntu'
    command = sys.argv[2]    # e.g., 'pytest tests/ -v'

    # Register cleanup in case it crashes
    atexit.register(lambda: cleanup(machine_name))
    
    # Start the VM
    run(f"vagrant up {machine_name}")

    # Run the user command inside the VM
    run(f"vagrant ssh {machine_name} -c \"{command}\"")
