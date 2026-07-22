fn main() -> Result<(), Box<dyn std::error::Error>> {
    let protoc = protoc_bin_vendored::protoc_bin_path()?;
    let includes = [
        std::path::PathBuf::from("proto"),
        protoc_bin_vendored::include_path()?,
    ];

    unsafe {
        std::env::set_var("PROTOC", protoc);
    }

    let protos = [std::path::PathBuf::from(
        "proto/cyberedge/v1/cyberedge.proto",
    )];

    tonic_prost_build::configure()
        .build_server(true)
        .build_client(true)
        .compile_protos(&protos, &includes)?;

    println!("cargo:rerun-if-changed=proto/cyberedge/v1/cyberedge.proto");
    Ok(())
}
