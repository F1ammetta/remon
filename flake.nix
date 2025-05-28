{
  description = "System Monitor - Remote systemd service management utility";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        
        # Build the Go application
        system-monitor = pkgs.buildGoModule {
          pname = "system-monitor";
          version = "1.0.0";
          
          src = ./.;
          
          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Will be updated by nix
          
          nativeBuildInputs = with pkgs; [
            templ
          ];
          
          preBuild = ''
            # Generate templ files
            templ generate
            
            # Copy static files to the build directory
            mkdir -p static
            cp static/* ./static/ 2>/dev/null || true
          '';
          
          postInstall = ''
            # Create directories for static files
            mkdir -p $out/share/system-monitor/static
            mkdir -p $out/share/system-monitor/templates
            
            # Copy static files
            cp -r static/* $out/share/system-monitor/static/ 2>/dev/null || true
            
            # Create wrapper script that sets the correct static file path
            mv $out/bin/system-monitor $out/bin/.system-monitor-unwrapped
            cat > $out/bin/system-monitor << 'EOF'
#!/bin/sh
export STATIC_PATH="${pkgs.lib.getBin system-monitor}/share/system-monitor/static"
cd "$STATIC_PATH/.."
exec ${pkgs.lib.getBin system-monitor}/bin/.system-monitor-unwrapped "$@"
EOF
            chmod +x $out/bin/system-monitor
          '';
          
          meta = with pkgs.lib; {
            description = "Remote systemd service management utility";
            homepage = "https://github.com/F1ammetta/remon";
            license = licenses.mit;
            maintainers = [ maintainers.yourusername ];
            platforms = platforms.linux;
          };
        };
        
        # Generate systemd service file
        systemd-service = pkgs.writeTextFile {
          name = "system-monitor.service";
          text = ''
            [Unit]
            Description=System Monitor - Remote systemd service management
            After=network.target
            Wants=network.target
            
            [Service]
            Type=simple
            User=system-monitor
            Group=system-monitor
            ExecStart=${system-monitor}/bin/system-monitor
            Environment=PORT=8080
            Environment=LOG_LEVEL=info
            Environment=MONITORED_SERVICES=nginx,postgresql,redis,docker,ssh,cron,systemd-resolved
            Environment=REFRESH_INTERVAL=30
            Environment=MAX_LOG_LINES=1000
            Restart=always
            RestartSec=5
            
            # Security settings
            NoNewPrivileges=true
            PrivateTmp=true
            ProtectSystem=strict
            ProtectHome=true
            ReadWritePaths=/var/lib/system-monitor
            
            # Allow systemctl commands
            SupplementaryGroups=wheel
            
            [Install]
            WantedBy=multi-user.target
          '';
        };
        
        # Installation script
        install-script = pkgs.writeShellScriptBin "install-system-monitor" ''
          set -e
          
          echo "🚀 Installing System Monitor..."
          
          # Create user and group
          if ! id system-monitor &>/dev/null; then
            echo "📝 Creating system-monitor user..."
            sudo useradd -r -s /bin/false -d /var/lib/system-monitor system-monitor
          fi
          
          # Create data directory
          echo "📁 Creating data directory..."
          sudo mkdir -p /var/lib/system-monitor
          sudo chown system-monitor:system-monitor /var/lib/system-monitor
          
          # Install binary
          echo "📦 Installing binary..."
          sudo cp ${system-monitor}/bin/system-monitor /usr/local/bin/
          sudo cp -r ${system-monitor}/share/system-monitor /usr/local/share/
          
          # Install systemd service
          echo "⚙️  Installing systemd service..."
          sudo cp ${systemd-service} /etc/systemd/system/system-monitor.service
          
          # Add system-monitor user to sudoers for systemctl
          echo "🔐 Configuring sudo permissions..."
          echo "system-monitor ALL=(ALL) NOPASSWD: /bin/systemctl" | sudo tee /etc/sudoers.d/system-monitor
          
          # Reload systemd and enable service
          echo "🔄 Enabling service..."
          sudo systemctl daemon-reload
          sudo systemctl enable system-monitor.service
          
          echo "✅ Installation complete!"
          echo ""
          echo "🎯 Next steps:"
          echo "  sudo systemctl start system-monitor    # Start the service"
          echo "  sudo systemctl status system-monitor   # Check status"
          echo "  journalctl -u system-monitor -f        # View logs"
          echo ""
          echo "🌐 The service will be available at http://localhost:8080"
        '';
        
        # Uninstall script
        uninstall-script = pkgs.writeShellScriptBin "uninstall-system-monitor" ''
          set -e
          
          echo "🗑️  Uninstalling System Monitor..."
          
          # Stop and disable service
          echo "⏹️  Stopping service..."
          sudo systemctl stop system-monitor.service || true
          sudo systemctl disable system-monitor.service || true
          
          # Remove files
          echo "📁 Removing files..."
          sudo rm -f /usr/local/bin/system-monitor
          sudo rm -rf /usr/local/share/system-monitor
          sudo rm -f /etc/systemd/system/system-monitor.service
          sudo rm -f /etc/sudoers.d/system-monitor
          
          # Remove user and group
          echo "👤 Removing user..."
          sudo userdel system-monitor || true
          sudo groupdel system-monitor || true
          
          # Remove data directory (optional)
          read -p "Remove data directory /var/lib/system-monitor? [y/N] " -n 1 -r
          echo
          if [[ $REPLY =~ ^[Yy]$ ]]; then
            sudo rm -rf /var/lib/system-monitor
          fi
          
          # Reload systemd
          sudo systemctl daemon-reload
          
          echo "✅ Uninstallation complete!"
        '';
        
      in
      {
        packages = {
          default = system-monitor;
          system-monitor = system-monitor;
          install = install-script;
          uninstall = uninstall-script;
        };
        
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go development
            go
            gopls
            gotools
            go-tools
            delve
            
            # Templ for template generation
            templ
            
            # Live reloading
            air
            
            # Git and development tools
            git
            gnumake
            
            # System tools
            systemd
            
            # Linting and formatting
            golangci-lint
            
            # Process management for development
            overmind
            tmux
          ];
          
          shellHook = ''
            echo "🚀 System Monitor Development Environment"
            echo "================================================"
            echo ""
            echo "📦 Available commands:"
            echo "  make dev              # Start development server with live reload"
            echo "  make build            # Build the application"
            echo "  make test             # Run tests"
            echo "  make templ-generate   # Generate templ files"
            echo "  make fmt              # Format code"
            echo "  make lint             # Lint code"
            echo ""
            echo "🔧 Development tools installed:"
            echo "  Go $(go version | cut -d' ' -f3)"
            echo "  Templ $(templ version 2>/dev/null || echo 'latest')"
            echo "  Air $(air -v 2>/dev/null || echo 'latest')"
            echo ""
            echo "🏗️  Installation commands (run outside nix shell):"
            echo "  nix run .#install     # Install system-wide"
            echo "  nix run .#uninstall   # Uninstall system-wide"
            echo ""
            
            # Create development directories
            mkdir -p tmp static
            
            # Copy static files if they don't exist
            if [ ! -f static/styles.css ]; then
              echo "📁 Setting up static files..."
              cp styles.css static/ 2>/dev/null || true
              cp script.js static/ 2>/dev/null || true
            fi
            
            # Set environment variables for development
            export PORT=8080
            export LOG_LEVEL=debug
            export MONITORED_SERVICES="nginx,postgresql,redis,docker,ssh,cron"
          '';
        };
        
        # NixOS module for easy integration
        nixosModules.default = { config, lib, pkgs, ... }:
          with lib;
          let
            cfg = config.services.system-monitor;
          in {
            options.services.system-monitor = {
              enable = mkEnableOption "System Monitor service";
              
              port = mkOption {
                type = types.int;
                default = 8080;
                description = "Port to listen on";
              };
              
              monitoredServices = mkOption {
                type = types.listOf types.str;
                default = [ "nginx" "postgresql" "redis" "docker" "ssh" "cron" ];
                description = "List of services to monitor";
              };
              
              refreshInterval = mkOption {
                type = types.int;
                default = 30;
                description = "Refresh interval in seconds";
              };
              
              maxLogLines = mkOption {
                type = types.int;
                default = 1000;
                description = "Maximum number of log lines to display";
              };
            };
            
            config = mkIf cfg.enable {
              systemd.services.system-monitor = {
                description = "System Monitor - Remote systemd service management";
                after = [ "network.target" ];
                wantedBy = [ "multi-user.target" ];
                
                environment = {
                  PORT = toString cfg.port;
                  LOG_LEVEL = "info";
                  MONITORED_SERVICES = concatStringsSep "," cfg.monitoredServices;
                  REFRESH_INTERVAL = toString cfg.refreshInterval;
                  MAX_LOG_LINES = toString cfg.maxLogLines;
                };
                
                serviceConfig = {
                  Type = "simple";
                  User = "system-monitor";
                  Group = "system-monitor";
                  ExecStart = "${system-monitor}/bin/system-monitor";
                  Restart = "always";
                  RestartSec = 5;
                  
                  # Security settings
                  NoNewPrivileges = true;
                  PrivateTmp = true;
                  ProtectSystem = "strict";
                  ProtectHome = true;
                  ReadWritePaths = [ "/var/lib/system-monitor" ];
                  SupplementaryGroups = [ "wheel" ];
                };
              };
              
              users.users.system-monitor = {
                isSystemUser = true;
                group = "system-monitor";
                home = "/var/lib/system-monitor";
                createHome = true;
              };
              
              users.groups.system-monitor = {};
              
              security.sudo.extraRules = [{
                users = [ "system-monitor" ];
                commands = [{
                  command = "${pkgs.systemd}/bin/systemctl";
                  options = [ "NOPASSWD" ];
                }];
              }];
            };
          };
      });
}

