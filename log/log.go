package log

import (
  "os"
  "fmt"
  "time"
  "bytes"
)

var LogsPath string

func combineLog(text, typz string) string {
  var buffer bytes.Buffer

  // Time layout string: "Mon Jan 2 15:04:05 MST 2006"
  buffer.WriteString(typz + " ")
  buffer.WriteString(time.Now().Format("02/01/2006 15:04:05 "))
  buffer.WriteString(text + "\n")

  return buffer.String()
}

func Log(text, typz string) error {
  var file, err = os.OpenFile(LogsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer file.Close()

  if err != nil {
    fmt.Println(err)
    return err
  }

  s := combineLog(text, typz)

  if _, err = file.WriteString(s); err != nil {
    fmt.Println(err)
    return err
  }

  file.Sync()
  return nil
}
