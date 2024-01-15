{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    let
      outputsWithoutSystem = { };
      outputsWithSystem = flake-utils.lib.eachDefaultSystem
        (system:
          let
            pkgs = import nixpkgs {
              inherit system;
            };
            lib = pkgs.lib;
          in
          {
            devShells = {
              default = pkgs.mkShell {
                buildInputs = with pkgs; [
                  go
                  mockgen
                  gopls
                  golangci-lint
                  wire
                  go-migrate
                  bunyan-rs
                  entr
                  bun
                ];
              };
            };
            packages = rec {
              default = docker;

              docker = pkgs.dockerTools.buildImage {
                name = "prismelabs/analytics";
                tag = "dev";

                copyToRoot = [ pkgs.cacert ];
                runAsRoot = ''
                  #!${pkgs.runtimeShell}
                  mkdir -p /app
                  cp -r ${self.packages."${system}".prisme-healthcheck}/bin/* /healthcheck
                '';

                config = {
                  Cmd = [ "${self.packages.${system}.prisme-bin}/bin/server" ];
                  WorkingDir = "/app";
                };
              };

              prisme-bin = pkgs.buildGoModule {
                pname = "prisme";
                version = "0.1.0";
                vendorHash = "sha256-sIqQerBGVuxk6Y5vhbGZA0OAR89OMKZ55Gsp0ippXrc=";

                src = ./.;
                # Skip go test.
                doCheck = false;
              };

              prisme-healthcheck = pkgs.writeShellApplication {
                name = "prisme-healthcheck";
                runtimeInputs = with pkgs; [ bash wget ];
                text = ''
                  wget --no-verbose --tries=1 --spider "http://localhost:''${PORT:-8000}/health_check" || exit 1
                '';
              };
            };
          });
    in
    outputsWithSystem // outputsWithoutSystem;
}
