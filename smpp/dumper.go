package smpp

import (
	"encoding/binary"
	"fmt"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/smpp/smpp/pdu"
	"github.com/mdouchement/smpp/smpp/pdu/pdufield"
	"github.com/mdouchement/smpp/smpp/pdu/pdutlv"
)

// Dump displays logs the given PDU data.
func Dump(l logger.Logger, p pdu.Body) {
	h := p.Header()

	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(h.ID))
	group := binary.BigEndian.Uint16(b[2:4])
	l = l.WithField("command_id", fmt.Sprintf("0x%08X", group))
	l = l.WithField("command_status", fmt.Sprintf("0x%X %s", uint32(h.Status), h.Status.Error()))
	l = l.WithField("sequence", h.Seq)

	// Mandatory fields (body)
	for k, v := range p.Fields() {
		l = l.WithField(string(k), field(v))
	}

	// Print optional fields (TLV)
	for k, v := range p.TLVFields() {
		l = l.WithField("tlv."+TagString(k), tlv(v))
	}

	l.WithPrefixf("[%s]", h.ID).Info("PDU")
}

func field(b pdufield.Body) string {
	if v, ok := b.(*pdufield.Fixed); ok {
		return fmt.Sprintf("0x%02X", v.Data)
	}

	return b.String()
}

func tlv(b pdutlv.Body) string {
	if len(b.Bytes()) == 1 {
		// Almost of the time an integer in that case.
		return fmt.Sprintf("0x%02X", b.Bytes()[0])
	}

	return b.String()
}

// TagString returns the string name of the given tag.
func TagString(t pdutlv.Tag) string {
	switch t {
	case pdutlv.TagDestAddrSubunit:
		return "dest_addr_subunit"
	case pdutlv.TagDestNetworkType:
		return "dest_network_type"
	case pdutlv.TagDestBearerType:
		return "dest_bearer_type"
	case pdutlv.TagDestTelematicsID:
		return "dest_telematics_id"
	case pdutlv.TagSourceAddrSubunit:
		return "source_addr_subunit"
	case pdutlv.TagSourceNetworkType:
		return "source_network_type"
	case pdutlv.TagSourceBearerType:
		return "source_bearer_type"
	case pdutlv.TagSourceTelematicsID:
		return "source_telematics_id"
	case pdutlv.TagQosTimeToLive:
		return "qos_time_to_live"
	case pdutlv.TagPayloadType:
		return "payload_type"
	case pdutlv.TagAdditionalStatusInfoText:
		return "additional_status_info_text"
	case pdutlv.TagReceiptedMessageID:
		return "receipted_message_id"
	case pdutlv.TagMsMsgWaitFacilities:
		return "ms_msg_wait_facilities"
	case pdutlv.TagScInterfaceVersion:
		return "sc_interface_version"
	case pdutlv.TagPrivacyIndicator:
		return "privacy_indicator"
	case pdutlv.TagSourceSubaddress:
		return "source_subaddress"
	case pdutlv.TagDestSubaddress:
		return "dest_subaddress"
	case pdutlv.TagUserMessageReference:
		return "user_message_reference"
	case pdutlv.TagUserResponseCode:
		return "user_response_code"
	case pdutlv.TagSourcePort:
		return "source_port"
	case pdutlv.TagDestinationPort:
		return "destination_port"
	case pdutlv.TagSarMsgRefNum:
		return "sar_msg_ref_num"
	case pdutlv.TagLanguageIndicator:
		return "language_indicator"
	case pdutlv.TagSarTotalSegments:
		return "sar_total_segments"
	case pdutlv.TagSarSegmentSeqnum:
		return "sar_segment_seqnum"
	case pdutlv.TagCallbackNumPresInd:
		return "callback_num_pres_ind"
	case pdutlv.TagCallbackNumAtag:
		return "callback_num_atag"
	case pdutlv.TagNumberOfMessages:
		return "number_of_messages"
	case pdutlv.TagCallbackNum:
		return "callback_num"
	case pdutlv.TagDpfResult:
		return "dpf_result"
	case pdutlv.TagSetDpf:
		return "set_dpf"
	case pdutlv.TagMsAvailabilityStatus:
		return "ms_availability_status"
	case pdutlv.TagNetworkErrorCode:
		return "network_error_code"
	case pdutlv.TagMessagePayload:
		return "message_payload"
	case pdutlv.TagDeliveryFailureReason:
		return "delivery_failure_reason"
	case pdutlv.TagMoreMessagesToSend:
		return "more_messages_to_send"
	case pdutlv.TagMessageStateOption:
		return "message_state_option"
	case pdutlv.TagUssdServiceOp:
		return "ussd_service_op"
	case pdutlv.TagDisplayTime:
		return "display_time"
	case pdutlv.TagSmsSignal:
		return "sms_signal"
	case pdutlv.TagMsValidity:
		return "ms_validity"
	case pdutlv.TagAlertOnMessageDelivery:
		return "alert_on_message_delivery"
	case pdutlv.TagItsReplyType:
		return "its_reply_type"
	case pdutlv.TagItsSessionInfo:
		return "its_session_info"
	default:
		return "0x" + t.Hex()
	}
}
