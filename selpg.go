/* ��ʵ�ó���ӱ�׼��������Ϊ�����в����������ļ�����ȡ�ı����롣
�������û�ָ�����Ը����벢��󽫱������ҳ�淶Χ�����磬�������
���� 100 ҳ�����û���ָ��ֻ��ӡ�� 35 �� 65 ҳ��*/
package main

/////////////////////////import//////////////////////////////// 

import (
	"io"
	"os/exec"
	"bufio"//bufio ������������ I/O ���档 
	 flag "github.com/spf13/pflag"
	"os"
	"fmt"
)

/////////////////////////selpg_args struct//////////////////////////////// 

type  selpg_args struct {
	start_page int  /* ��ӡ[start_page, end_page]������� */
	end_page int
	
	in_filename string  /* �����ļ����� */ 
	print_dest string	/* ����ļ��� */
	
	page_len int  /* ÿҳ��������Ĭ��Ϊ72 */
	page_type string  /* 'l'���д�ӡ��'f'����ҳ����ӡ��Ĭ�ϰ��� */
}

type sp_args selpg_args /* ������ */

/////////////////////////global variable//////////////////////////////// 

var progname string /* �������ƣ��������ͨ�������Ʊ����ã���ȫ�ֱ�������Ϊ�ڴ�����Ϣ����ʾ֮�� */

/////////////////////////main//////////////////////////////// 

func main() {
	sa := sp_args{}
	progname = os.Args[0]
	
	// ������� 
	process_args(&sa)
	// ����������� 
	process_input(sa)
}

/////////////////////////func process_args//////////////////////////////// 

func process_args(sa * sp_args) {
	/*��flag�󶨵�sa�ĸ��������� */ 
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
	
	/* ��������кϷ��� */
	/* os.Args��һ�����������в�����string���飬���ǿ���ʹ���±������ʲ��� */
	
	if len(os.Args) < 3 {	/* ������������������Ϊ progname -s start_page -e end_page�� */
		fmt.Fprintf(os.Stderr, "\n%s: not enough arguments\n", progname)
		flag.Usage()
		os.Exit(1)
	}

	/* �����һ������ - start page */
	/* ��һ����������Ϊ's'��start_page �������1����С�ڼ�����ܱ�ʾ���������ֵ */ 
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

	/* ����ڶ������� - end page */
	/* ��һ����������Ϊ'e'��end_page �������1��С�ڼ�����ܱ�ʾ���������ֵ����С�ڵ���start_page*/ 
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

	/* ����page_len */
	if ( sa.page_len < 1 || sa.page_len > (INT_MAX - 1) ) {
		fmt.Fprintf(os.Stderr, "\n%s: invalid page length %s\n", progname, sa.page_len)
		flag.Usage()
		os.Exit(5)
	}
	
	/* ����in_filename */ 
	/*����Ƿ���ʣ��Ĳ��������� selpg�������һ�������Ĳ�������������������ļ�����*/
	if len(flag.Args()) == 1 {
		_, err := os.Stat(flag.Args()[0])
		/* ����ļ��Ƿ���� */
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
	/* ������ */
	var fin *os.File 
	
	/* ����������������������� */
	
	/* ������������������������նˣ��û����̣����ļ�����һ���������� */
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
	//ʹ�� bufio.NewReader �����һ����ȡ������
	bufFin := bufio.NewReader(fin)
	
	
	/* ��������ص㡣�����������Ļ���ļ�����һ���ļ������� */
	
	var fout io.WriteCloser
	/*������������cmd��ʹ�÷���������������*/
	cmd := &exec.Cmd{}

	if len(sa.print_dest) == 0 {
		fout = os.Stdout
	} else {
		cmd = exec.Command("cat")
		
		//��ֻд�ķ�ʽ�� print_dest �ļ�������ļ������ڣ��ʹ������ļ��� 
		var err error
		cmd.Stdout, err = os.OpenFile(sa.print_dest, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s: could not open file %s\n",
				progname, sa.print_dest)
			os.Exit(8)
		}
		
		//StdinPipe����һ�����ӵ�command��׼����Ĺܵ�pipe 
		fout, err = cmd.StdinPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s: could not open pipe to \"lp -d%s\"\n",
				progname, sa.print_dest)
			os.Exit(8)
		}
		
		cmd.Start()
		defer fout.Close()
	}


	/* ��������ӡ������ */
	
	/* ����page_type�����̶��������ҳ�����д�ӡ�� */
	
	//��ǰҳ�� 
	var page_ctr int

	if sa.page_type == "l" { //���̶�������ӡ 
		line_ctr := 0
		page_ctr = 1
		for {
			/*����д����bufFin := bufio.NewReader(fin)*/
			line,  crc := bufFin.ReadString('\n')
			if crc != nil {
				break // ����eof 
			}
			line_ctr++
			if line_ctr > sa.page_len {
				page_ctr++
				line_ctr = 1
			}
			/*����ָ��ҳ�룬��ʼ��ӡ*/ 
			if (page_ctr >= sa.start_page) && (page_ctr <= sa.end_page) {
				_, err := fout.Write([]byte(line))
				if err != nil {
					fmt.Println(err)
					os.Exit(9)
				}
		 	}
		}  
	} else {			//����ҳ����ӡ 
		page_ctr = 1
		for {
			page, err := bufFin.ReadString('\n')
			//txt û�л�ҳ����ʹ��\n���棬���ұ��ڲ���
			//line, crc := bufFin.ReadString('\f')
			if err != nil {
				break // eof
			}
			/*����ָ��ҳ�룬��ʼ��ӡ*/ 
			if (page_ctr >= sa.start_page) && (page_ctr <= sa.end_page) {
				_, err := fout.Write([]byte(page))
				if err != nil {
					os.Exit(5)
				}
			}
			//ÿ����һ����ҳ��������һҳ 
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