package main

import (
	"bufio"
	"io"
	"os"
	"time"
)

type ReadableClient struct {
	onDataHandler func(bytes []byte)
	file *os.File
	onEndHandler func()
	statusReading bool
}
func Readable(filePath string) *ReadableClient{
	readerStep := make([]byte, 65536)
	r := &ReadableClient{
		statusReading: true,
	}
	r.file, _ = os.Open(filePath)
	if r.file != nil {
		reader := bufio.NewReader(r.file)
		go func() {
			for {
				if r.statusReading && r.onDataHandler != nil{
					n, err := reader.Read(readerStep)
					if err != nil{
						if err == io.EOF {
							if r.onEndHandler != nil{
								r.onEndHandler()
							}
							break
						}
					}
					r.onDataHandler(readerStep[:n])
				}
				time.Sleep(time.Millisecond * 10)
			}
		}()
	}else{
		return nil
	}
	return r
}
func (r *ReadableClient) onData(handler func(bytes []byte)) {
	r.onDataHandler = handler
}
func (r *ReadableClient) onEnd(handler func())  {
	r.onEndHandler = handler
}
func (r *ReadableClient) pause(){
	r.statusReading = false
}
func (r *ReadableClient) resume(){
	r.statusReading = true
}
func (r *ReadableClient) getFilesizeInBytes() int {
	fi, err := r.file.Stat()
	if err != nil {
		return 0
	}
	return normalizeInt(fi.Size())
}