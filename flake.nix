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
              vendorHash = "sha256-oxr+DxEB9C6h/X8z61HTMISzv//Fv2YBf1V2jnrnonA=";

              src = ./.;
              # Skip go test.
              doCheck = false;

              tags = tags;

              subPackages = "./cmd/prisme";
            };
          buildPrismeDocker =
            prisme: extraEnv:
            pkgs.dockerTools.buildImage {
              name = "prismelabs/analytics";
              tag = "dev";

              copyToRoot = [ pkgs.cacert ];
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
                  mockgen # Go mock generator
                  gopls # Go LSP
                  golangci-lint # Go linter
                  go-migrate # Go SQL migration
                  bunyan-rs # Bunyan format pretty print
                  entr # FS watcher
                  bun # Bun JS runtime
                  minify # JS minifier
                  clickhouse # clickhouse client
                ])
                ++ (with gengeommdb.packages.${system}; [ default ])
                ++ (with self.packages.${system}; [ libchdb ]);

              LD_LIBRARY_PATH = "${lib.makeLibraryPath buildInputs}";
            };
          };
          packages = rec {
            default = prisme;

            prisme = buildPrisme [ ];
            prisme-chdb = buildPrisme [ "chdb" ];

            docker = buildPrismeDocker prisme [ ];
            docker-chdb = buildPrismeDocker prisme-chdb [
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
