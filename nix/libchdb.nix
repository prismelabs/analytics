{
  stdenv,
  lib,
  fetchzip,
}:

stdenv.mkDerivation rec {
  pname = "libchdb";
  version = "3.7.2";

  outputs = [ "out" ];

  arch =
    if stdenv.isx86_64 then
      "x86_64"
    else if stdenv.isAarch64 then
      "arm64"
    else
      lib.fatalError "unsupported CPU architecture";

  src = fetchzip {
    url = "https://github.com/chdb-io/chdb/releases/download/v${version}/linux-${arch}-libchdb.tar.gz";
    sha256 = "sha256-GCTaHOW46owrPj4UhD8pREUBylPAiFiouXGR+la/UbQ=";
    stripRoot = false;
  };

  installPhase = ''
    mkdir -p "$out"/lib "$out"/include
    cp libchdb.so $out/lib/libchdb.so
    cp chdb.h chdb.hpp $out/include/
  '';
}
