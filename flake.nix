{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    gengeommdb.url = "github:negrel/gengeommdb";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      gengeommdb,
      ...
    }:
    let
      outputsWithoutSystem = { };
      outputsWithSystem = flake-utils.lib.eachDefaultSystem (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
          lib = pkgs.lib;
          buildPrisme =
            tags:
            pkgs.buildGoModule {
              pname = "prisme";
              version = "0.20.0";
              vendorHash = "sha256-Ab3eWUFnT4fTX6MR1MHdw0lpBlQmASrOG3/08vKQTrA=";

              src = ./.;
              # Skip go test.
              doCheck = false;

              tags = tags;

              subPackages = "./cmd/prisme";
            };
          dockerBuildPrisme =
            prisme: extraEnv:
            pkgs.dockerTools.buildImage {
              name = "prismelabs/analytics";
              tag = "dev";

              copyToRoot = with pkgs; [
                cacert
                bash
                coreutils
                bunyan-rs
              ];
              runAsRoot = ''
                #!${pkgs.runtimeShell}
                mkdir -p /app
                cp -r ${self.packages."${system}".prisme-healthcheck}/bin/* /healthcheck
              '';

              config = {
                Cmd = [ "${prisme}/bin/prisme" ];
                WorkingDir = "/app";
                Env = [
                  "PRISME_ADMIN_HOSTPORT=0.0.0.0:9090"
                ]
                ++ extraEnv;
              };
            };
        in
        {
          pkgs = pkgs;
          lib = lib;
          devShells = {
            default = pkgs.mkShell rec {
              buildInputs =
                (with pkgs; [
                  go
                  air # FS watcher / Live reload
                  gopls # Go LSP
                  golangci-lint # Go linter
                  go-migrate # Go SQL migration
                  bunyan-rs # Bunyan format pretty print
                  bun # Bun JS runtime
                  minify # JS minifier
                  clickhouse # clickhouse client
                  yq # Command-line YAML/XML/TOML processor
                  python313Packages.openapi-spec-validator # OpenAPI validator
                ])
                ++ (with gengeommdb.packages.${system}; [ default ])
                ++ (with self.packages.${system}; [ libchdb ]);

              LD_LIBRARY_PATH = "${lib.makeLibraryPath buildInputs}";
              GOFLAGS = "-tags=chdb";
            };
          };
          packages = rec {
            default = prisme;

            prisme = buildPrisme [ ];
            prisme-chdb = buildPrisme [ "chdb" ];

            docker = dockerBuildPrisme prisme [ ];
            docker-chdb = dockerBuildPrisme prisme-chdb [
              "CHDB_LIB_PATH=${libchdb}/lib/libchdb.so"
            ];

            prisme-healthcheck = pkgs.writeShellApplication {
              name = "prisme-healthcheck";
              runtimeInputs = with pkgs; [ wget ];
              text = ''
                wget --no-verbose --tries=1 --spider "http://localhost:''${PRISME_PORT:-80}/api/v1/healthcheck" || exit 1
              '';
            };

            libchdb = pkgs.callPackage ./nix/libchdb.nix { };
          };
        }
      );
    in
    outputsWithSystem // outputsWithoutSystem;
}
