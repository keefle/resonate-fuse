run_fuse ()
{
    if [ -e fake ]; then
        fusermount -u fake
        rm -rf fake
    fi

    if [ -e real ]; then
        rm -rf real
    fi

    mkdir real
    mkdir fake

    go run . 2>&1 | tee fuse.log

    fusermount -u fake
    rm -rf fake
    rm -rf real
}

run_fuse
