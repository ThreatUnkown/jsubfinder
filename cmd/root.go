package cmd

import (
	"errors"
	"os"

	C "github.com/ThreatUnkown/jsubfinder/core"
	l "github.com/ThreatUnkown/jsubfinder/core/logger"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "JSubFinder <Command> <Flags>",
		Short: "jsubfinder searches webpages for javascript & analyzes them for hidden subdomains and secrets (wip).",
		Long:  `jsubfinder searches webpages for javascript & analyzes them for hidden subdomains and secrets (wip).`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
	}
)

func init() {
	l.Log = logrus.New()
	rootCmd.AddCommand(searchExec)
	rootCmd.AddCommand(proxyExec)

	rootCmd.PersistentFlags().StringVarP(&C.OutputFile, "output-file", "o", "", "name/location to store the file")
	rootCmd.PersistentFlags().StringVarP(&C.SecretsOutputFile, "secrets", "s", "secrets.txt", "name/location to store the secrets, using --secret=\"\" will use the default filename")
	rootCmd.PersistentFlags().BoolVarP(&C.Debug, "debug", "d", false, "Enable debug mode. Logs are stored in log.info")

	rootCmd.PersistentFlags().BoolVarP(&C.Silent, "silent", "S", false, "Disable printing to the console")
	rootCmd.PersistentFlags().StringVar(&C.Sig, "sig", "", "Location of signatures for finding secrets")
	rootCmd.PersistentFlags().BoolVarP(&C.SSL, "nossl", "K", true, "Skip SSL cert verification")

}

func Execute() error {
	return rootCmd.Execute()
}

//Things to check before running any code.
func safetyChecks() error {

	if C.Debug && C.Silent { //if debug and silent flag are enabled return error
		l.Log.SetLevel(logrus.DebugLevel)
		return errors.New("please choose Debug mode or silent mode. Enabling both is conflicting")
	} else if C.Debug { //Setup logging

		l.InitDetailedLogger()
		l.Log.SetLevel(logrus.DebugLevel)
		l.Log.Debug("Debug mode enabled")

	} else if C.Silent { //If silent supress all errors
		l.Log.SetLevel(logrus.PanicLevel)
	}

	if C.Silent && C.OutputFile == "" { //if silent and no output return error
		l.Log.SetLevel(logrus.DebugLevel)
		return errors.New("please disable silent mode or output the results to a file with the -o flag otherwise you can't view the results")
	}

	//Check if we can write to the outputFile
	if C.OutputFile != "" {
		file, err := os.OpenFile(C.OutputFile, os.O_WRONLY, 0666)
		if err != nil {
			if os.IsPermission(err) {
				return (err)
			}

		}
		file.Close()
	}

	//Ensure you don't provide both url and input file
	if C.InputFile != "" && len(C.InputURLs) != 0 {
		return errors.New("Provide either -f or -u, you can't provide both")
	}

	//ensure signature file exists
	if rootCmd.Flags().Changed("secrets") {
		C.FindSecrets = true

		if C.Sig == "" {
			home, err := homedir.Dir()
			if err != nil {
				l.Log.Debug("Unable to find homedir, please provide the location of the signature fil with -sig")
				return err
			}
			C.Sig = home + "/.jsf_signatures.yaml"
		}

		//Load signatures for secrets
		err := C.ConfigSigs.ParseConfig(C.Sig) //https://github.com/eth0izzle/shhgit/blob/090e3586ee089f013659e02be16fd0472b629bc7/core/signatures.go
		if err != nil {
			return err
		}
		C.Signatures, err = C.ConfigSigs.GetSignatures()
		if err != nil {
			return err
		}
		C.Blacklisted_extensions = []string{".exe", ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif", ".psd", ".xcf", ".zip", ".tar.gz", ".ttf", ".lock"}

		if C.Silent == true {
			C.PrintSecrets = false
		} else {
			C.PrintSecrets = true
		}

		if C.SecretsOutputFile == "" {
			C.SecretsOutputFile = "secrets.txt"
		}

	}

	//Ensure output and secret file aren't the same
	if C.OutputFile == C.SecretsOutputFile {
		return errors.New("The output file for the results and secrets can't be the same")
	}

	//if silent && debug {
	//	l.Log.Fatal("Enable silent mode or debug mode. Can't print debug information if silent mode is enabled.")
	//}

	//ensure output is being sent to console or outputfile.
	if C.Silent && C.OutputFile == "" {
		return errors.New("if you aren't saving the output with -o and you want the console output silenced -S, what's the point of running JSubfinder?")
	}

	return nil
}
