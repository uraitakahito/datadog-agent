name "sds"

# default_version "4112a1c"
default_version "main"

source git: 'https://github.com/DataDog/dd-sensitive-data-scanner'

relative_path "sds-go/rust"

build do
    license "Apache-2.0"
    license_file "./LICENSE"

    if linux_target?
        command "ls -lah"
        command "cargo build --release"
        command "cp target/release/libsds_go.so #{install_dir}/embedded/lib"
    end
end
