package nex

import (
	"os"
	"strconv"

	"github.com/PretendoNetwork/friends/globals"
	nex "github.com/PretendoNetwork/nex-go"
	ticket_granting "github.com/PretendoNetwork/nex-protocols-go/ticket-granting"
	common_ticket_granting "github.com/PretendoNetwork/nex-protocols-common-go/ticket-granting"
)

func registerCommonAuthenticationServerProtocols() {
	ticketGrantingProtocol := ticket_granting.NewProtocol(globals.AuthenticationServer)
	commonTicketGrantingProtocol := common_ticket_granting.NewCommonTicketGrantingProtocol(ticketGrantingProtocol)

	port, _ := strconv.Atoi(os.Getenv("PN_FRIENDS_SECURE_SERVER_PORT"))

	secureStationURL := nex.NewStationURL("")
	secureStationURL.Scheme = "prudps"
	secureStationURL.Fields.Set("address", os.Getenv("PN_FRIENDS_SECURE_SERVER_HOST"))
	secureStationURL.Fields.Set("port", strconv.Itoa(port))
	secureStationURL.Fields.Set("CID", "1")
	secureStationURL.Fields.Set("PID", "2")
	secureStationURL.Fields.Set("sid", "1")
	secureStationURL.Fields.Set("stream", "10")
	secureStationURL.Fields.Set("type", "2")

	commonTicketGrantingProtocol.SecureStationURL = secureStationURL
	commonTicketGrantingProtocol.BuildName = serverBuildString
	commonTicketGrantingProtocol.EnableInsecureLogin()

	globals.AuthenticationServer.SetPasswordFromPIDFunction(globals.PasswordFromPID)
}
