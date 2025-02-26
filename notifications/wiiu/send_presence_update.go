package notifications_wiiu

import (
	"github.com/PretendoNetwork/friends/database"
	database_wiiu "github.com/PretendoNetwork/friends/database/wiiu"
	"github.com/PretendoNetwork/friends/globals"
	nex "github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/constants"
	"github.com/PretendoNetwork/nex-go/v2/types"
	friends_wiiu_types "github.com/PretendoNetwork/nex-protocols-go/v2/friends-wiiu/types"
	nintendo_notifications "github.com/PretendoNetwork/nex-protocols-go/v2/nintendo-notifications"
	nintendo_notifications_types "github.com/PretendoNetwork/nex-protocols-go/v2/nintendo-notifications/types"
)

func SendPresenceUpdate(presence friends_wiiu_types.NintendoPresenceV2) {
	eventObject := nintendo_notifications_types.NewNintendoNotificationEvent()
	eventObject.Type = types.NewUInt32(24)
	eventObject.SenderPID = presence.PID.Copy().(types.PID)
	eventObject.DataHolder = types.NewDataHolder()
	eventObject.DataHolder.Object = presence.Copy().(friends_wiiu_types.NintendoPresenceV2)

	stream := nex.NewByteStreamOut(globals.SecureEndpoint.LibraryVersions(), globals.SecureEndpoint.ByteStreamSettings())

	eventObject.WriteTo(stream)

	notificationRequest := nex.NewRMCRequest(globals.SecureEndpoint)
	notificationRequest.ProtocolID = nintendo_notifications.ProtocolID
	notificationRequest.CallID = 3810693103
	notificationRequest.MethodID = nintendo_notifications.MethodProcessNintendoNotificationEvent2
	notificationRequest.Parameters = stream.Bytes()

	notificationRequestBytes := notificationRequest.Bytes()

	friendList, err := database_wiiu.GetUserFriendList(uint32(presence.PID))
	if err != nil && err != database.ErrEmptyList {
		globals.Logger.Critical(err.Error())
	}

	if friendList == nil {
		return
	}

	// TODO - Not needed?
	friends := friendList.Copy().(types.List[friends_wiiu_types.FriendInfo])

	for i := 0; i < len(friends); i++ {
		friend := friends[i]

		/*
			if friend == nil || friend.NNAInfo == nil || friend.NNAInfo.PrincipalBasicInfo == nil {
				// TODO - Fix this
				pid := presence.PID
				var friendPID uint32 = 0

				if friend != nil && friend.Presence != nil {
					// TODO - Better track the bad users PID
					friendPID = friend.Presence.PID
				}

				globals.Logger.Error(fmt.Sprintf("User %d has friend %d with bad presence data", pid, friendPID))

				if friend == nil {
					globals.Logger.Error(fmt.Sprintf("%d friendList[i] nil", friendPID))
				} else if friend.NNAInfo == nil {
					globals.Logger.Error(fmt.Sprintf("%d friendList[i].NNAInfo is nil", friendPID))
				} else if friend.NNAInfo.PrincipalBasicInfo == nil {
					globals.Logger.Error(fmt.Sprintf("%d friendList[i].NNAInfo.PrincipalBasicInfo is nil", friendPID))
				}

				continue
			}
		*/

		// * This is meant to replace the commented out check.
		// * The check was meant to catch cases where friend
		// * presences would be nil, but now we no longer use
		// * pointers so that can't ever happen. Instead, check
		// * for the default value(s)
		if friend.NNAInfo.PrincipalBasicInfo.PID == 0 {
			continue
		}

		friendPID := friend.NNAInfo.PrincipalBasicInfo.PID
		connectedUser, ok := globals.ConnectedUsers.Get(uint32(friendPID))

		if ok && connectedUser != nil {
			requestPacket, _ := nex.NewPRUDPPacketV0(globals.SecureEndpoint.Server, connectedUser.Connection, nil)

			requestPacket.SetType(constants.DataPacket)
			requestPacket.AddFlag(constants.PacketFlagNeedsAck)
			requestPacket.AddFlag(constants.PacketFlagReliable)
			requestPacket.SetSourceVirtualPortStreamType(connectedUser.Connection.StreamType)
			requestPacket.SetSourceVirtualPortStreamID(globals.SecureEndpoint.StreamID)
			requestPacket.SetDestinationVirtualPortStreamType(connectedUser.Connection.StreamType)
			requestPacket.SetDestinationVirtualPortStreamID(connectedUser.Connection.StreamID)
			requestPacket.SetPayload(notificationRequestBytes)

			globals.SecureServer.Send(requestPacket)
		}
	}
}
