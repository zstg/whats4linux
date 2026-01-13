{
  lib,
  version ? "0.0.1",
  makeWrapper,
  buildGoModule,
  pkg-config,
  wails,
  gtk3,
  pango,
  glib,
  harfbuzz,
  atk,
  cairo,
  gdk-pixbuf,
  zlib,
  fontconfig,
  webkitgtk_4_1,
  libsoup_3,
  nodejs_latest,
  fetchNpmDeps,
  npmHooks,
}:

buildGoModule {
  pname = "whats4linux";
  inherit version;
  
  src = ./.;
  
  vendorHash = "sha256-T1SEsdG+aFnfa0jpwhooOXJu/bzhSAAyDx49fD466V4=";
  
  proxyVendor = true;
  
  npmDeps = fetchNpmDeps {
    inherit version;
    src = ./frontend;
    hash = "sha256-QbMNbPKmWFaBszM2CCj6Qrd2b602K5z4zXgrqLJESmk=";
  };
  
  subPackages = [ ]; # this defaults to "."
  doBuild = false; 
  
  tags = [ "desktop,production" ];
  
  nativeBuildInputs = [
    makeWrapper
    pkg-config
    wails
    nodejs_latest
    npmHooks.npmConfigHook
  ];
  
  buildInputs = [
    gtk3.dev
    pango.dev
    glib.dev
    harfbuzz.dev
    atk.dev
    cairo.dev
    gdk-pixbuf.dev
    zlib.dev
    fontconfig.dev
    webkitgtk_4_1.dev
    libsoup_3.dev
  ];
  
  postPatch = ''
    # Copy package.json and package-lock.json to root for npmConfigHook to find them
    cp frontend/package.json ./
    cp frontend/package-lock.json ./
  '';
  
  preBuild = ''
    # Set up a proper home directory for binding generation
    export HOME=$(mktemp -d)
    
    # Build with Wails - it will automatically handle frontend build
    wails build -tags "webkit2_41,soup_3"
  '';

  buildPhase = "runHook preBuild"; # no `go build`

  installPhase = ''
    runHook preInstall

    mkdir -p $out/bin $out/lib
    [ -d build/bin ] && cp build/bin/whats4linux $out/bin/ || true   
    [ -d build/lib ] && cp build/lib/* $out/lib/ || true   

    runHook postInstall
  '';
  
  postFixup = ''
    # Wrap the binary with required library paths
    wrapProgram $out/bin/whats4linux \
      --prefix LD_LIBRARY_PATH : "${lib.makeLibraryPath [
        gtk3
        webkitgtk_4_1
        libsoup_3
      ]}" \
      --prefix PKG_CONFIG_PATH : "${lib.concatStringsSep ":" [
        "${gtk3.dev}/lib/pkgconfig"
        "${webkitgtk_4_1.dev}/lib/pkgconfig"
        "${pango.dev}/lib/pkgconfig"
        "${glib.dev}/lib/pkgconfig"
        "${harfbuzz.dev}/lib/pkgconfig"
        "${atk.dev}/lib/pkgconfig"
        "${cairo.dev}/lib/pkgconfig"
        "${gdk-pixbuf.dev}/lib/pkgconfig"
        "${libsoup_3.dev}/lib/pkgconfig"
        "${zlib.dev}/lib/pkgconfig"
        "${fontconfig.dev}/lib/pkgconfig"
      ]}"
  '';

  passthru.updateScript = "./update.sh --update-vendor";
    
  meta = {
    homepage = "https://github.com/lugvitc/whats4linux";
    description = "An unofficial WhatsApp client for Linux";
    license = lib.licenses.agpl3Plus;
    maintainers = with lib.maintainers; [
      zstg
    ];
    platforms = lib.platforms.linux;
    mainProgram = "whats4linux";
  };
}