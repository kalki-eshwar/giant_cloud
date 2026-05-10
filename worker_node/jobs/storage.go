package jobs

import (
    "os"
    "io"
)

// writeFile persists the full payload to disk at the specified path.
func writeFile(path string, data []byte) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    n, err := f.Write(data)
    if err != nil {
        f.Close()
        return err
    }
    if n < len(data) {
        f.Close()
        return io.ErrShortWrite
    }
    f.Close()
    return nil
}
