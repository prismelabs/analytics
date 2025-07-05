{
  stdenv,
  lib,
  fetchzip,
}:

stdenv.mkDerivation rec {
  pname = "libchdb";
  version = "3.4.0";

  outputs = [ "out" ];

  arch =
    if stdenv.isx86_64 then
      "x86_64"
    else if stdenv.isAarch64 then
      "arm64"
    else
      lib.fatalError "unsupported CPU architecture";

  src = fetchzip {
    url = "https://github.com/chdb-io/chdb/releases/download/v${version}/linux-x86_64-libchdb.tar.gz";
    sha256 = "sha256:1b29g97f7d23y5ycdfgirm1lhlyv81lg0ffc4jbw9rrizzj6k4d5";
    stripRoot = false;
  };

  installPhase = ''
    mkdir -p "$out"/lib
    cp libchdb.so $out/lib/libchdb.so
  '';
}
