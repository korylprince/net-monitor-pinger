package icmpv4

//checksum calculates the [Internet Checksum](https://tools.ietf.org/html/rfc1071)
func checksum(packet []byte) uint16 {
	var sum uint32

	//pad to 16bit words
	if len(packet)%2 != 0 {
		packet = append(packet, 0)
	}

	words := len(packet) / 2

	for i := 0; i < words; i++ {
		word := uint32(packet[i*2])<<8 | uint32(packet[i*2+1])

		sum += word

		//carry one if present
		sum = (0xFFFF & sum) + (sum >> 16)
	}
	return ^uint16(sum)
}
