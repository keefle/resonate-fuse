# API Calles in resonatefuse


## Read
* Inputs:
    * bytes  data
    * offset int64
* Outputs:
    * error

## ReadAll
* Inputs:
    * bytes data
* Outputs:
    * err   error

## ReadDirAll
* Inputs:
    * NULL
* Outputs:
    * dirs fuse.Dirents
    * err error

## Readlink
* Inputs:
    * NULL
* Ouputs:
    * target string
    * err error

## Lookup
* Inputs:
    * name string
* Ouputs:
    * file File
    * err  error

## Create
* Inputs:
    * name string
    * mode fileMode
* Ouputs:
    * file File
    * error err


## Write
* Inputs:
    * data  bytes
    * offset int64
* Ouputs:
    * written int
    * err error

## Remove
* Inputs:
    * name string
* Ouputs:
    * err error

## Rename
* Inputs:
    * oldName string
    * newName string
    * newDir  File
* Ouputs:
    * err error

## Mkdir
* Inputs:
    * name string
    * mode fileMode
* Ouputs:
    * file File
    * err error

## Link
* Inputs:
    * newName string
    * old File
* Ouputs:
    * file File
    * err error

## Symlink
* Inputs:
    * target string
    * newName string
* Ouputs:
    * file FFile
    * err error

## Setattr
* Inputs:
    * mode  fileMode
    * atime time
    * mtime time
* Ouputs:
    * err error
