run_fuse () {
    go run . &> fuse.log &
}

test_dir() {
    if [ -e /tmp/test_dir ]; then
        rm -rf /tmp/test_dir
    fi

    mkdir /tmp/test_dir/
    cd /tmp/test_dir/
    echo "{ \"name\": \"test\" }" > package.json
    npm install node-ios-device
    find ./ -type f -exec md5sum {} + | grep -v 'package.json' | sort -k 2 > /tmp/bench.txt
    rm -rf /tmp/test_dir
}

buildup() {
    if [ -e fake ]; then
        fusermount -u fake
        rm -rf fake
    fi

    if [ -e real ]; then
        rm -rf real
    fi

    mkdir real
    mkdir fake
}

cleanup() {
    fusermount -u fake
    rm -rf fake real
}

fake_dir() {
    buildup
    run_fuse
    sleep 1
    (
        cd fake
        echo "{ \"name\": \"test\" }" > package.json
        npm install node-ios-device
        find ./ -type f -exec md5sum {} + | grep -v 'package.json' | sort -k 2 > /tmp/fake.txt
    )
    pkill example
    cleanup
}




main() {
    fake_dir
    test_dir

    echo "diffing"
    diff /tmp/bench.txt /tmp/fake.txt
    sha256sum /tmp/bench.txt
    sha256sum /tmp/fake.txt
    rm /tmp/bench.txt /tmp/fake.txt
}

main
