{
  inputs = { flake-utils.url = "github:numtide/flake-utils"; };

  outputs = { self, nixpkgs, flake-utils, ... }:
    let
      outputsWithoutSystem = { };
      outputsWithSystem = flake-utils.lib.eachDefaultSystem (system:
        let
          pkgs = import nixpkgs { inherit system; };
          lib = pkgs.lib;
          libraryPath = lib.makeLibraryPath [ self.packages.${system}.chdb ];
        in {
          libraryPath = libraryPath;
          devShells = {
            default = pkgs.mkShell {
              buildInputs = (with pkgs; [
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
              ]) ++ (with self.packages.${system}; [ chdb ]);

              LD_LIBRARY_PATH = libraryPath;
            };
          };
          packages = rec {
            default = pkgs.buildGoModule {
              pname = "prisme";
              version = "0.18.0";
              vendorHash =
                "sha256-TV5KN7IIkTswlBjLXwpdYxDycHu5wOknIyaa/o7DzVw=";

              src = ./.;
              # Skip go test.
              doCheck = false;

              postBuild = ''
                mv "$GOPATH/bin/server" "$GOPATH/bin/prisme"
              '';

              subPackages = "./cmd/server";

              ldflags = [ "-extldflags '-L${chdb}/lib'" ];
            };

            docker = pkgs.dockerTools.buildImage {
              name = "prismelabs/analytics";
              tag = "dev";

              copyToRoot = [ pkgs.cacert ];
              runAsRoot = ''
                #!${pkgs.runtimeShell}
                mkdir -p /app
                cp -r ${
                  self.packages."${system}".prisme-healthcheck
                }/bin/* /healthcheck
              '';

              config = {
                Cmd = [ "${self.packages.${system}.default}/bin/prisme" ];
                WorkingDir = "/app";
                Env = [
                  "PRISME_ADMIN_HOSTPORT=0.0.0.0:9090"
                  "LD_LIBRARY_PATH=${libraryPath}"
                ];
              };
            };

            prisme-healthcheck = pkgs.writeShellApplication {
              name = "prisme-healthcheck";
              runtimeInputs = with pkgs; [ wget ];
              text = ''
                wget --no-verbose --tries=1 --spider "http://localhost:''${PRISME_PORT:-80}/api/v1/healthcheck" || exit 1
              '';
            };

            chdb = pkgs.stdenv.mkDerivation rec {
              name = "chdb";
              version = "v2.1.1";

              src = builtins.fetchTarball {
                url =
                  "https://github.com/chdb-io/chdb/releases/download/${version}/linux-x86_64-libchdb.tar.gz";
                sha256 =
                  "sha256:1jy53li55v3vkfkakx05rjfgjw0sghl471im31fnnys7xigk8mly";
              };

              buildPhase = ''
                mkdir -p $out/include $out/lib
                cp chdb.h $out/include
                cp libchdb.so $out/lib
              '';
            };
          };
        });
    in outputsWithSystem // outputsWithoutSystem;
}
