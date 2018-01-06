package wordfilter

import (
	"bufio"
	"bytes"
	"container/list"
	"io"
	"os"
)

func appendStringByByte(w *string, p byte) {
	tmp := []byte{}
	tmp = []byte(*w)
	tmp = append(tmp, p)
	//for i := 0; i < len(tmp); i++ {
	//	if tmp[i] == 0 {
	//		*w = string(tmp[0:i])
	//		return
	//	}
	//}
	*w = string(tmp)
}

func ReNameFile(srcdir string) error {
	bakdir := srcdir + "_bak"

	_, err := os.Stat(bakdir)
	if err == nil {
		os.Remove(bakdir)
	}

	return os.Rename(srcdir, bakdir)
}

func ReadByLineMap(dir string) (map[string]uint32, error) {
	f, err := os.OpenFile(dir, os.O_CREATE, 0666)
	if err != nil {
		panic(err)
		return nil, err
	}
	defer f.Close()

	lines := make(map[string]uint32, 0)
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadBytes('\n')
		//line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		if err != nil || io.EOF == err {
			break
		}
		if len(line) == 0 || line[0] == '#' || line[0] == ';' {
			continue
		}
		line = bytes.Replace(line, []byte("\n"), nil, -1)
		line = bytes.Replace(line, []byte("\r"), nil, -1)
		lines[string(line)] = 1

	}

	return lines, err
}

func WriteByLine(dir string, contents map[string]uint32) error {
	f, err := os.OpenFile(dir, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f) //创建新的 Writer 对象
	for c, _ := range contents {
		_, err = w.WriteString(c + "\n")
		if err != nil {
			panic(err)
			continue
		}
	}
	w.Flush()
	return nil
}

func ReadByLine(dir string) (*list.List, error) {
	f, err := os.Open(dir)
	if err != nil {
		panic(err)
		return nil, err
	}
	defer f.Close()

	lines := list.New()
	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadBytes('\n')
		//line, err := rd.ReadString('\n') //以'\n'为结束符读入一行
		if err != nil || io.EOF == err {
			break
		}
		if len(line) == 0 || line[0] == '#' || line[0] == ';' {
			continue
		}
		line = bytes.Replace(line, []byte("\n"), nil, -1)
		line = bytes.Replace(line, []byte("\r"), nil, -1)
		lines.PushBack(string(line))

	}

	return lines, err
}
