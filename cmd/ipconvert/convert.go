package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"
)

//go:embed qip
var qipexecutable []byte

func convertIpv4(qip string) error {
	fmt.Print("convert ipv4.awdb to neo.ipv4.ipdb ...\n")
	cmd := exec.Command(qip, "pack", "-s", "ipv4.awdb", "-l", "ipline.awdb", "--ipipFile", "ipv4.ipdb", "-f", "country,province,city,owner_domain,isp,latitude,longitude,timezone,utc_offset,china_admin_code,idd_code,country_code,continent_code|country=中国:country,province,city,,isp,,,,,,,areacode,continent_code|country,,,,,,,,,,,areacode,continent_code", "-o", "neo.ipv4.ipdb")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func convertIpv6(qip string) error {
	fmt.Print("convert ipv6.awdb to neo.ipv6.ipdb ...\n")

	cmd := exec.Command(qip, "pack", "-s", "ipv6.awdb", "-f", "country,province,city,owner_domain,isp,latitude,longitude,timezone,utc_offset,china_admin_code,idd_code,country_code,continent_code|country=中国:country,province,city,,isp,,,,,,,areacode,continent_code|country,,,,,,,,,,,areacode,continent_code", "-o", "neo.ipv6.ipdb")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func awdb2ipdb(ctx context.Context, executable string) error {
	wg, ctx := errgroup.WithContext(ctx)

	for _, doConvert := range []func(string) error{convertIpv4, convertIpv6} {
		wg.Go(func() error {
			return doConvert(executable)
		})
	}
	return wg.Wait()
}

func prepareExecutable(filename string) error {
	fp, err := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_TRUNC, 0700)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrExist):
			os.Remove(filename)
			return prepareExecutable(filename)
		default:
			return fmt.Errorf("os.OpenFile: %w", err)
		}
	}
	fp.Write(qipexecutable)
	return fp.Close()
}
