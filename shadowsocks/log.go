package shadowsocks

import (
	"io"
	"log"
	"os"
	"ss_go/utils"
)

var g_logger *log.Logger = nil

func init() {
	logPath := utils.GetMainDirectory() + utils.GetExeFileName() + ".log"
	log.Printf("log path is %s.", logPath)
	fout, err := os.OpenFile(logPath, os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to open log file:%s.", logPath)
		return
	}

	// 多重输出
	writers := []io.Writer{
		fout,
		os.Stdout,
	}
	multiWriter := io.MultiWriter(writers...)

	//
	g_logger = log.New(multiWriter, "", log.LstdFlags | log.Lshortfile)
}

//
func LOG_INFO_F(format string, args ...interface{}) {
	if g_logger == nil {
		return
	}
	g_logger.SetPrefix("[info] ")
	g_logger.Printf(format, args ...)
}

//
func LOG_INFO(args ...interface{}) {
	if g_logger == nil {
		return
	}
	g_logger.SetPrefix("[info] ")
	g_logger.Println(args ...)
}

//
func LOG_ERROR_F(format string, args ...interface{}) {
	if g_logger == nil {
		return
	}
	g_logger.SetPrefix("[error] ")
	g_logger.Printf(format, args ...)
}

//
func LOG_ERROR(args ...interface{}) {
	if g_logger == nil {
		return
	}
	g_logger.SetPrefix("[error] ")
	g_logger.Println(args ...)
}