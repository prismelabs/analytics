{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    let
      outputsWithoutSystem = { };
      outputsWithSystem = flake-utils.lib.eachDefaultSystem
        (system:
          let
            pkgs = import nixpkgs {
              inherit system;
            };
            # lib = pkgs.lib;
          in
          {
            devShells = {
              default = pkgs.mkShell {
                buildInputs = with pkgs; [
                  go
                  mockgen # Go mock generator
                  gopls # Go LSP
                  golangci-lint # Go linter
                  wire # Go dependency injection
                  go-migrate # Go SQL migration
                  bunyan-rs # Bunyan format pretty print
                  entr # FS watcher
                  bun # Bun JS runtime
                  minify # JS minifier
                  clickhouse # clickhouse client
                  hyperfine # binary benchmarks
                ];
              };
            };
            packages = {
              default = pkgs.buildGoModule {
                pname = "prisme";
                version = "0.15.0";
                vendorHash = "sha256-ILeTewHQGd41xcS9RTIoKKid6pOOKMuGb5fiCl09v2w=";

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
                  Env = [ "PRISME_ADMIN_HOSTPORT=0.0.0.0:9090" ];
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
