/*
 * Created by Sumesh K K on 12/31/2016.
 * This code will build an exe file which can read a YAML file with the set up instruction to install the packages
 * usage instruction: InstallSw.exe --install softwareNames_separated_by_comma
Eg: InstallSw.exe --install cntlm,maven
 */
package main
 
import ("fmt"
"io/ioutil"
"os"
"os/exec"
"github.com/cavaliercoder/grab"
"gopkg.in/yaml.v2"
"strings"
"log"
"flag"
"time"
 
)
 
var directoryForDownloadingSoftware string
var instalaltionCommandline string
var swPackageExtension string
var logFilePath string
var swPostInstalaltionCommandline string
var swPostInstallationBanner string
var configurationFileUrl string = "http://cdtsdvo108p.rxcorp.com:8421/repository/DevOps/softwareConfiguration.yaml"
var unzipUrl string = "http://cdtsdvo108p.rxcorp.com:8421/repository/DevOps/unzip.exe"
var unzipFilePath string
var swName string
var swUrl string
var swCommandline string
func main() {
 
    unzipChannel := make(chan string)
    yamlChannel := make(chan string)
    otherChannel := make(chan string)
    softwareNameChannel :=make(chan string)
    softwareUrlChannel := make(chan string)
    swCommandlineChannel:=make(chan string)
    swPostInstalaltionCommandlineChannel:=make(chan string)
    swPostInstallationBannerChannel :=make(chan string)
    //set the folder to download all packages . Get the Userprofile environment variable
 
    var profilePath string = os.Getenv("USERPROFILE")
 
    commandInputArrayFlag := flag.String("install","","List of software comma separated Eg: cntlm,sbt")
    flag.Parse()
    directoryForDownloadingSoftware = profilePath + "\\devopsPackage"
    logFilePath = directoryForDownloadingSoftware+"\\Installation.log"
    fmt.Println("Starting installation and setup: Please find the logs at",logFilePath)
    logFile,_ :=os.OpenFile(logFilePath,os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
    defer logFile.Close()
    log.SetOutput(logFile)
    log.Println("Starting")
    log.Println(directoryForDownloadingSoftware)
    err := os.Mkdir(directoryForDownloadingSoftware, 0755)
    if err != nil {
 
        log.Println("Allready exists: ",directoryForDownloadingSoftware,"Deleting all files from the directory")
        os.RemoveAll(directoryForDownloadingSoftware)
 
    }
    //Download the Yaml and unzip.exe which are prequisites
 
    log.Println("Dowloading configuration YAML file from  ",configurationFileUrl, "To",directoryForDownloadingSoftware)
    go downloadAFile(configurationFileUrl, directoryForDownloadingSoftware, yamlChannel)
    log.Println("Dowloading Unzip file from  ",unzipUrl, "To",directoryForDownloadingSoftware)
    go downloadAFile(unzipUrl, directoryForDownloadingSoftware, unzipChannel)
    unzipFilePath := <-unzipChannel
    configFilePath := <-yamlChannel
        log.Println("Configuration FilePath is:",configFilePath)
    log.Println("Unzip file is :",unzipFilePath)
    log.Println("Commandline array is :",*commandInputArrayFlag)
    for _,commandLineInputCaseSensitive := range (strings.Split(*commandInputArrayFlag,",")){
        //making sure that strings are transformed as lower case
        commandLineInput :=strings.ToLower(commandLineInputCaseSensitive)
        log.Println("Parsing configuration file for and returning the attributes for ",commandLineInput)
        //Parse YAML file get the software url and other details
        go parseYaml(commandLineInput, configFilePath,softwareNameChannel, softwareUrlChannel ,swCommandlineChannel,swPostInstalaltionCommandlineChannel,swPostInstallationBannerChannel)
        softwareName := <-softwareNameChannel
        softwareUrl :=<-softwareUrlChannel
        swCommandline :=<-swCommandlineChannel
        swPostInstalaltionCommandline:=<-swPostInstalaltionCommandlineChannel
        swPostInstallationBanner:=<-swPostInstallationBannerChannel
        log.Println("Software Name: ", softwareName,"Software Url:", softwareUrl,"Commandline:", swCommandline)
        if softwareUrl == "" {
            log.Printf("Matching software %s not found in configuration file : Exiting Now",softwareName)
            panic("Didnt find any matching software with the name")
        }
                log.Printf("Downloading %s to %s",softwareUrl,directoryForDownloadingSoftware)
        //download software from software Url mentioned in the yaml file
        go downloadAFile(softwareUrl, directoryForDownloadingSoftware, otherChannel)
        downloadedFilePath := <-otherChannel
        splitStringArray := strings.Split(softwareUrl,".")
        swPackageExtension=splitStringArray[len(splitStringArray)-1]
        fmt.Println("Starting setup of",downloadedFilePath)
        log.Printf("Starting setup of",downloadedFilePath)
        log.Printf("Running setup %s %s %s %s %s %s",downloadedFilePath,swCommandline,unzipFilePath,swPackageExtension)
        //start installation
        Installation(downloadedFilePath,swCommandline,unzipFilePath,swPackageExtension,swPostInstalaltionCommandline,swPostInstallationBanner)
        log.Println("Setup completed for ",downloadedFilePath)
        fmt.Println("Setup completed for ")
    }
}
 
func parseYaml(softwareName string, configFilePath string,softwareNameChannel chan string, softwareUrlChannel chan string,swCommandlineChannel chan string,swPostInstalaltionCommandlineChannel chan string,swPostInstallationBannerChannel chan string) {
 
 
    //} (url string,installationType string){
    type yamlStruct []struct {
        Name             string `yaml:"Name"`
        Url              string `yaml:"Url"`
        InstallationCommandline    string `yaml:"InstallationCommandline"`
        PostInstallationCommandline string `yaml:"PostInstallationCommandline"`
        ConfigurationFileUrl string    `yaml:"ConfigurationFileUrl"`
        PostInstallationBanner string    `yaml:"PostInstallationBanner"`
 
 
    }
 
    openFile, err := ioutil.ReadFile(configFilePath)
    if err != nil {
        panic(fmt.Sprintf("Stopping execution %s", err))
 
    }
    var configFileStructure yamlStruct
    yaml.Unmarshal(openFile, &configFileStructure)
    log.Println(configFileStructure)
    for _, value := range configFileStructure {
        if value.Name == softwareName {
            swName = value.Name
            swUrl = value.Url
            swCommandline = value.InstallationCommandline
            swPostInstalaltionCommandline =value.PostInstallationCommandline
            swPostInstallationBanner = value.PostInstallationBanner
            break
 
        }
 
    }
    softwareNameChannel<-softwareName
    softwareUrlChannel<-swUrl
    swCommandlineChannel<-swCommandline
    swPostInstalaltionCommandlineChannel<-swPostInstalaltionCommandline
    swPostInstallationBannerChannel<-swPostInstallationBanner
 
}
 
func downloadAFile(sourceUrl string, destinationPath string, downloadFileChannel chan string) {
 
    fmt.Println("Download of",sourceUrl,"started")
    log.Println("Download of",sourceUrl,"started")
    resp, err := grab.Get(destinationPath, sourceUrl)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", destinationPath, sourceUrl)
        log.Println("Error in downloading")
        os.Exit(1)
    }
 
    fmt.Printf("Successfully downloaded %s to %s\n",sourceUrl, resp.Filename)
    log.Println("Dowload completed ,downloaded file Name is ",resp.Filename)
    var downloadedFileName string = resp.Filename
    downloadFileChannel <- downloadedFileName
    log.Println("Transferring the data to download Channel")
}
 
func Installation(downloadedFilePath string,swCommandline string, unzipFilePath ,swPackageExtension string,postInstallationCommandLine string,
postInstallationBanner string) {
    switch swPackageExtension {
    case "exe":
 
            log.Println("Package type: exe")
            fmt.Println("Executing setup")
            exec.Command(downloadedFilePath).Run()
        defer func() {
            if err := recover(); err != nil {
                fmt.Println("Panic handled locally")
                log.Println("Panic handled locally")
            }
        }()
        case "msi":
            log.Println("Package type: msi")
            fmt.Println("Executing setup")
            err :=exec.Command("cmd","/c",downloadedFilePath).Run()
            if err != nil {
                log.Println("Error occured :",err,"Exiting")
                panic(err)
            }
 
 
        case "zip":
            log.Println("Package type: Unzip")
            err := exec.Command(unzipFilePath, downloadedFilePath, "-d", directoryForDownloadingSoftware).Run()
            fmt.Println("Exec dommandline", unzipFilePath, downloadedFilePath, directoryForDownloadingSoftware)
            log.Println("Exec dommandline", unzipFilePath, downloadedFilePath, directoryForDownloadingSoftware)
            if err != nil {
            panic(err)
        } else {
            if swCommandline != "None" {
 
                fmt.Println("Commandline is", swCommandline)
                instalaltionCommandline = directoryForDownloadingSoftware + swCommandline
                fmt.Println("Installation command is", instalaltionCommandline)
                log.Println("Installation command is", instalaltionCommandline)
                err = exec.Command("cmd", "/c", instalaltionCommandline).Run()
 
                if err:=recover();err != nil {
                    fmt.Println("If you are trying to install Babun then make sure .babun directory is deleted from Your profile folder")
                    log.Println("If you are trying to install Babun then make sure .babun directory is deleted from Your profile folder :Exiting")
                    panic(err)
                }
 
            }
        }
 
    }
fmt.Println("Completed setup..")
    fmt.Println("Installation Completed Running any post installation commands")
    log.Println("Installation Completed Running any post installation commands")
        PostInstallation(postInstallationCommandLine,postInstallationBanner)
}
 
func PostInstallation(postInstalaltionCommandLine string,postInstallationBanner string){
 
    switch postInstalaltionCommandLine {
 
    case "None":
        break
    default:
        err := exec.Command("cmd","/c",postInstalaltionCommandLine).Run()
        if err != nil{
            log.Println("Post installation command failed",postInstalaltionCommandLine)
        }
        log.Println("Displaying post instalaltion banner ",postInstallationBanner)
        fmt.Println(postInstallationBanner)
        break
    }
 
    time.Sleep(5*time.Second)
 
}
