package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"mvdan.cc/xurls"
)

func main() {

	port := flag.String("port", "80", "port to forward")
	host := flag.String("host", "ferreteria.cifu.dev", "host to be replaced")
	apacheVersion := flag.String("apachev", "2.4.54.2", "apache httpd version of wamp")

	flag.Parse()

	subDomain := strings.Split(*host, ".")

	if len(subDomain) == 0 {
		panic("did not find subdomain in host parameter")
	}

	wpConfig := fmt.Sprintf("C:\\wamp64\\www\\%s\\wp-config.php", subDomain[0])
	vHosts := fmt.Sprintf("C:\\wamp64\\bin\\apache\\apache%s\\conf\\extra\\httpd-vhosts.conf", *apacheVersion)

	fmt.Println()

	funnelCmdOutput, err := execute("tailscale", "funnel", "--bg", string(*port))
	if err != nil {
		panic(err)
	}

	rxRelaxed := xurls.Relaxed
	funnelUrlString := rxRelaxed.FindString(funnelCmdOutput)

	funnelUrl, err := url.Parse(funnelUrlString)
	if err != nil {
		panic(err)
	}

	err = replace_text_in_file(wpConfig, *host, funnelUrl.Host)
	if err != nil {
		panic(err)
	}

	err = replace_text_in_file(vHosts, *host, funnelUrl.Host)
	if err != nil {
		panic(err)
	}

	fmt.Println()

	err = reiniciar_apache()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Comparte la URL  %s", funnelUrlString)
	fmt.Println()
	fmt.Println()
	fmt.Printf("Presiona Enter para cerrar el túnel y revertir los cambios del config (apache y wordpress)")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	_, err = reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	_, err = execute("tailscale", "funnel", "reset")
	if err != nil {
		panic("Problem with reset")
	}

	err = replace_text_in_file(wpConfig, funnelUrl.Host, *host)
	if err != nil {
		panic(err)
	}

	err = replace_text_in_file(vHosts, funnelUrl.Host, *host)
	if err != nil {
		panic(err)
	}

	err = reiniciar_apache()
	if err != nil {
		panic(err)
	}

	fmt.Println("Listo")
}

func reiniciar_apache() error {
	_, err := execute("net", "stop", "wampapache64")
	if err != nil {
		return err
	}

	_, err = execute("net", "start", "wampapache64")
	if err != nil {
		return err
	}

	return nil

}

func execute(program string, args ...string) (string, error) {

	fmt.Printf("Ejecutando %s with args: %s", program, strings.Join(args, " "))
	fmt.Println()

	cmd := exec.Command(program, args...)
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	fmt.Println(string(stdout))
	fmt.Println()

	return string(stdout), nil
}

func replace_text_in_file(file_path string, find, replace string) error {

	read, err := os.ReadFile(file_path)
	if err != nil {
		return err
	}

	file_name := filepath.Base(file_path)

	fmt.Println()
	fmt.Printf("Reemplazando en archivo %s el texto '%s' por '%s'", file_name, find, replace)
	fmt.Println()

	newContents := strings.Replace(string(read), find, replace, -1)

	err = os.WriteFile(file_path, []byte(newContents), 0)
	if err != nil {
		return err
	}

	return nil
}
