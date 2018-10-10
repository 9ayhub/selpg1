/* 该实用程序从标准输入或从作为命令行参数给出的文件名读取文本输入。
它允许用户指定来自该输入并随后将被输出的页面范围。例如，如果输入
含有 100 页，则用户可指定只打印第 35 至 65 页。*/
package main

/////////////////////////import//////////////////////////////// 

import (
	"io"
	"os/exec"
	"bufio"//bufio 用来帮助处理 I/O 缓存。 
	 flag "github.com/spf13/pflag"
	"os"
	"fmt"
)

/////////////////////////selpg_args struct//////////////////////////////// 

type  selpg_args struct {
	start_page int  /* 打印[start_page, end_page]间的内容 */
	end_page int
	
	in_filename string  /* 输入文件名， */ 
	print_dest string	/* 输出文件名 */
	
	page_len int  /* 每页的行数，默认为72 */
	page_type string  /* 'l'按行打印，'f'按换页符打印，默认按行 */
}

type sp_args selpg_args /* 简化名字 */

/////////////////////////global variable//////////////////////////////// 

var progname string /* 保存名称（命令就是通过该名称被调用）的全局变量，作为在错误消息中显示之用 */

/////////////////////////main//////////////////////////////// 

func main() {
	sa := sp_args{}
	progname = os.Args[0]
	
	// 处理参数 
	process_args(&sa)
	// 处理输入输出 
	process_input(sa)
}

/////////////////////////func process_args//////////////////////////////// 

func process_args(sa * sp_args) {
	/*将flag绑定到sa的各个变量上 */ 
	flag.IntVarP(&sa.start_page,"start",  "s", -1, "start page(>1)")
	flag.IntVarP(&sa.end_page,"end", "e",  -1, "end page(>=start_page)")
	flag.IntVarP(&sa.page_len,"len", "l", 72, "page len")
	flag.StringVarP(&sa.print_dest,"dest", "d", "", "print dest")
	flag.StringVarP(&sa.page_type,"type", "f", "l", "'l' for lines-delimited, 'f' for form-feed-delimited. default is 'l'")
	flag.Lookup("type").NoOptDefVal = "f"
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"USAGE: \n%s -s start_page -e end_page [ -f | -l lines_per_page ]" + 
			" [ -d dest ] [ in_filename ]\n", progname)
		flag.PrintDefaults()
	}
	
	flag.Parse()
	
	/* 检查命令行合法性 */
	/* os.Args是一个储存了所有参数的string数组，我们可以使用下标来访问参数 */
	
	if len(os.Args) < 3 {	/* 参数个数不够（至少为 progname -s start_page -e end_page） */
		fmt.Fprintf(os.Stderr, "\n%s: not enough arguments\n", progname)
		flag.Usage()
		os.Exit(1)
	}

	/* 处理第一个参数 - start page */
	/* 第一个参数必须为's'，start_page 必须大于1，且小于计算机能表示的最大整数值 */ 
	if os.Args[1] != "-s" {
		fmt.Fprintf(os.Stderr, "\n%s: 1st arg should be -s start_page\n", progname)
		flag.Usage()
		os.Exit(2)
	}
	
	INT_MAX := 1 << 32 - 1
	
	if(sa.start_page < 1 || sa.start_page > INT_MAX) {
		fmt.Fprintf(os.Stderr, "\n%s: invalid start page %s\n", progname, os.Args[2])
		flag.Usage()
		os.Exit(3)
	}

	/* 处理第二个参数 - end page */
	/* 第一个参数必须为'e'，end_page 必须大于1，小于计算机能表示的最大整数值，且小于等于start_page*/ 
	if os.Args[3] != "-e" {
		fmt.Fprintf(os.Stderr, "\n%s: 2nd arg should be -e end_page\n", progname)
		flag.Usage()
		os.Exit(4)
	}
	
	if sa.end_page < 1 || sa.end_page > INT_MAX || sa.end_page < sa.start_page {
		fmt.Fprintf(os.Stderr, "\n%s: invalid end page %s\n", progname, sa.end_page)
		flag.Usage()
		os.Exit(5)
	}

	/* 处理page_len */
	if ( sa.page_len < 1 || sa.page_len > (INT_MAX - 1) ) {
		fmt.Fprintf(os.Stderr, "\n%s: invalid page length %s\n", progname, sa.page_len)
		flag.Usage()
		os.Exit(5)
	}
	
	/* 设置in_filename */ 
	/*检查是否还有剩余的参数。对于 selpg，最多有一个这样的参数，它被用作输入的文件名。*/
	if len(flag.Args()) == 1 {
		_, err := os.Stat(flag.Args()[0])
		/* 检查文件是否存在 */
		if err != nil && os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "\n%s: input file \"%s\" does not exist\n",
					progname, flag.Args()[0]);
			os.Exit(6);
		}
		
		sa.in_filename = flag.Args()[0]
	}

	/* page_len */ 
}

/////////////////////////func process_input//////////////////////////////// 

func process_input(sa sp_args) {
	/* 输入流 */
	var fin *os.File 
	
	/* ···设置输入输出··· */
	
	/* 设置输入流。输入可以来自终端（用户键盘），文件或另一个程序的输出 */
	if len(sa.in_filename) == 0 {
		fin = os.Stdin
	} else {
		var err error
		fin, err = os.Open(sa.in_filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s: could not open input file \"%s\"\n",
				progname, sa.in_filename)
			os.Exit(7)
		}
		defer fin.Close()
	}
	//使用 bufio.NewReader 来获得一个读取器变量
	bufFin := bufio.NewReader(fin)
	
	
	/* 设置输出地点。输出可以是屏幕，文件或另一个文件的输入 */
	
	var fout io.WriteCloser
	/*？？？？？？cmd的使用方法？？？？？？*/
	cmd := &exec.Cmd{}

	if len(sa.print_dest) == 0 {
		fout = os.Stdout
	} else {
		cmd = exec.Command("cat")
		
		//用只写的方式打开 print_dest 文件，如果文件不存在，就创建该文件。 
		var err error
		cmd.Stdout, err = os.OpenFile(sa.print_dest, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s: could not open file %s\n",
				progname, sa.print_dest)
			os.Exit(8)
		}
		
		//StdinPipe返回一个连接到command标准输入的管道pipe 
		fout, err = cmd.StdinPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s: could not open pipe to \"lp -d%s\"\n",
				progname, sa.print_dest)
			os.Exit(8)
		}
		
		cmd.Start()
		defer fout.Close()
	}


	/* ···打印··· */
	
	/* 根据page_type（按固定行数或分页符进行打印） */
	
	//当前页数 
	var page_ctr int

	if sa.page_type == "l" { //按固定行数打印 
		line_ctr := 0
		page_ctr = 1
		for {
			/*上文写到的bufFin := bufio.NewReader(fin)*/
			line,  crc := bufFin.ReadString('\n')
			if crc != nil {
				break // 碰到eof 
			}
			line_ctr++
			if line_ctr > sa.page_len {
				page_ctr++
				line_ctr = 1
			}
			/*到达指定页码，开始打印*/ 
			if (page_ctr >= sa.start_page) && (page_ctr <= sa.end_page) {
				_, err := fout.Write([]byte(line))
				if err != nil {
					fmt.Println(err)
					os.Exit(9)
				}
		 	}
		}  
	} else {			//按分页符打印 
		page_ctr = 1
		for {
			page, err := bufFin.ReadString('\n')
			//txt 没有换页符，使用\n代替，而且便于测试
			//line, crc := bufFin.ReadString('\f')
			if err != nil {
				break // eof
			}
			/*到达指定页码，开始打印*/ 
			if (page_ctr >= sa.start_page) && (page_ctr <= sa.end_page) {
				_, err := fout.Write([]byte(page))
				if err != nil {
					os.Exit(5)
				}
			}
			//每碰到一个换页符都增加一页 
			page_ctr++
		}
	}
	
	//if err := cmd.Wait(); err != nil {
		//handle err
		if page_ctr < sa.start_page {
			fmt.Fprintf(os.Stderr,
				"\n%s: start_page (%d) greater than total pages (%d)," +
				" no output written\n", progname, sa.start_page, page_ctr)
		} else if page_ctr < sa.end_page {
			fmt.Fprintf(os.Stderr,"\n%s: end_page (%d) greater than total pages (%d)," +
			" less output than expected\n", progname, sa.end_page, page_ctr)
		} /*else {
			fmt.Fprintf(os.Stderr,"\n%s: cmd.Start() failed with %s\n", progname, err)
		} */
	/*} else {
		fmt.Fprintf(os.Stderr,"\n%s: done! \n", progname)
	}*/
	
	
}