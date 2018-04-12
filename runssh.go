package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"golang.org/x/crypto/ssh"
	"strconv"
	"strings"
	"time"
)

type host struct {
	host, port, user, passwd string
}

type cmd struct {
	hidx, cidx int
	err        error
	data       string
	errout     string
}

func readX(inputFile *string) ([]host, []cmd, []string, error) {
	var hosts []host
	var commands []cmd
	var cmdstrings []string

	xlsx, err := excelize.OpenFile(*inputFile)
	if err != nil {
		return hosts, commands, cmdstrings, err
	}
	sheetName := xlsx.GetSheetName(1)
	if sheetName == "" {
		return hosts, commands, cmdstrings, errors.New("Empty workbook")
	}
	rows := xlsx.GetRows(sheetName)
	if len(rows) < 1 {
		return hosts, commands, cmdstrings, errors.New("Empty sheet")
	}
	if len(rows[0]) < 5 {
		return hosts, commands, cmdstrings, errors.New("Need at least 1 command")
	}
	var ih, iu, ipw, ip int

	for i, c := range rows[0][0:4] {
		switch strings.ToLower(c) {
		case "host":
			ih = i
		case "username":
			iu = i
		case "password":
			ipw = i
		case "port":
			ip = i
		case "user":
			iu = i
		}
	}
	for i, r := range rows[1:] {
		var h host
		h.host = r[ih]
		h.user = r[iu]
		h.passwd = r[ipw]
		h.port = r[ip]
		hosts = append(hosts, h)
		for j, cr := range rows[0][4:] {
			var c cmd
			c.hidx = i
			c.cidx = j
			c.data = cr
			if i == 0 {
				cmdstrings = append(cmdstrings, cr)
			}
			commands = append(commands, c)
		}
	}
	return hosts, commands, cmdstrings, nil
}

func runSSH(h *host, c *cmd) *cmd {
	sshConfig := &ssh.ClientConfig{
		User: h.user,
		Auth: []ssh.AuthMethod{ssh.Password(h.passwd)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	conn, err := ssh.Dial("tcp", h.host+":"+h.port, sshConfig)
	if err != nil {
		c.err = err
		return c
	}
	session, err := conn.NewSession()
	defer session.Close()
	if err != nil {
		c.err = err
		return c
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf
	session.Run(c.data)
	c.data = stdoutBuf.String()
	c.errout = stderrBuf.String()
	c.err = nil
	return c
}

func main() {
	inputFile := flag.String("input", "", "Input file")
	outputFile := flag.String("output", "", "Output file, will overwrite input if not specified")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Input file not specified")
		flag.PrintDefaults()
	}
	if *outputFile == "" {
		*outputFile = *inputFile
	}
	hosts, commands, cmdstrings, err := readX(inputFile)
	if err != nil {
		fmt.Println("Unable to read input file:", err)
		return
	}

	results := make(chan *cmd, 10)
	timeout := time.After(30 * time.Second)
	for _, c := range commands {
		c2 := c
		go func(h *host, c *cmd) {
			results <- runSSH(h, c)
		}(&hosts[c2.hidx], &c2)
	}

	var xlsx *excelize.File
	if *inputFile == *outputFile {
		xlsx, err = excelize.OpenFile(*outputFile)
		if err != nil {
			fmt.Println("Unable to open file")
			return
		}
	} else {
		xlsx = excelize.NewFile()
	}
	if xlsx == nil {
		fmt.Println("Unable to create new spreadsheet")
		return
	}

readResults:
	for i := 0; i < len(commands); i++ {
		select {
		case c := <-results:
			cs := excelize.ToAlphaString(c.cidx+4) + strconv.Itoa(c.hidx+2)
			xlsx.SetCellValue("Sheet1", cs, c.data)
		case <-timeout:
			fmt.Println("Timeout")
			break readResults
		}
	}
	if *inputFile != *outputFile {
		for i, h := range hosts {
			ch := excelize.ToAlphaString(0) + strconv.Itoa(i+2)
			cp := excelize.ToAlphaString(1) + strconv.Itoa(i+2)
			cu := excelize.ToAlphaString(2) + strconv.Itoa(i+2)
			cpw := excelize.ToAlphaString(3) + strconv.Itoa(i+2)
			xlsx.SetCellValue("Sheet1", ch, h.host)
			xlsx.SetCellValue("Sheet1", cp, h.port)
			xlsx.SetCellValue("Sheet1", cu, h.user)
			xlsx.SetCellValue("Sheet1", cpw, h.passwd)
		}
		xlsx.SetCellValue("Sheet1", "A1", "Host")
		xlsx.SetCellValue("Sheet1", "B1", "Port")
		xlsx.SetCellValue("Sheet1", "C1", "User")
		xlsx.SetCellValue("Sheet1", "D1", "Password")
		for i, s := range cmdstrings {
			cc := excelize.ToAlphaString(4+i) + "1"
			xlsx.SetCellValue("Sheet1", cc, s)
		}
	}
	err = xlsx.SaveAs(*outputFile)
	if err != nil {
		fmt.Println("Error saving output file")
	}
}
