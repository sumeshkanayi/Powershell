$ErrorActionPreference="stop"
Add-Type -Path 'Renci.SshNet.dll'
$routerList=import-csv .\input.csv
$date = $((Get-Date).ToString('yyyy-MM-dd-HHmm')) 
$userNmae="root"
$password="root123"

foreach($router in $routerList)
{

$deviceName=$router.routerName
$deviceIp=$router.ip

$commandToExecute="show running-config"
#$commandToExecute="ifconfig"

$SshClient = New-Object Renci.SshNet.SshClient($deviceIp, 22, $userNmae, $password)


try{
"Connecting to $deviceName"
$SshClient.Connect()
}
catch{

"Error connecting to router $deviceName Continuing"

"Error connecting to router $deviceName Continuing" |ac error.log

continue

}

if ($SshClient.IsConnected) {
    $SshCommand = $SshClient.RunCommand($commandToExecute)		# Result of 'ifconfig' is returned to $SshCommand
    $output = $SshCommand.Result.Split("`n")			# Split up the result into individual lines for easier parsing
    $filename =$deviceName  + "_" + $date+ '.txt'
    $output | ac $filename #Change this to suit your environment
    "exported configuration to $filename"

 
}

}

'***************************'
"completed"