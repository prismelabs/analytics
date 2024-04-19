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
                  minify
                ];
              };
            };
            packages = rec {
              default = pkgs.buildGoModule {
                pname = "prisme";
                version = "0.14.0";
                vendorHash = "sha256-BJuvB3/5f3qu3icMREy6VBuHAZJ2iMbvYEUFxZInmCg=";

                src = ./.;
                # Skip go test.
                doCheck = false;

                postBuild = ''
                  mv "$GOPATH/bin/server" "$GOPATH/bin/prisme"
                '';

                subPackages = "./cmd/server";
              };

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
                  Cmd = [ "${self.packages.${system}.default}/bin/prisme" ];
                  WorkingDir = "/app";
                };
              };

              prisme-healthcheck = pkgs.writeShellApplication {
                name = "prisme-healthcheck";
                runtimeInputs = with pkgs; [ wget ];
                text = ''
                  wget --no-verbose --tries=1 --spider "http://localhost:''${PRISME_PORT:-80}/api/v1/healthcheck" || exit 1
                '';
              };
            };
          });
    in
    outputsWithSystem // outputsWithoutSystem;
}
