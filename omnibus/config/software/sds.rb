name "sds"

build do
    license "Apache-2.0"
    source git: 'https://github.com/DataDog/dd-sensitive-data-scanner.git'

    relative_path "dd-sensitive-data-scanner/sds-go/rust"

    if linux_target?
        command "cargo build --release"
        command "cp target/release/libsds_go.so #{install_dir}/embedded/lib"
    end
end
