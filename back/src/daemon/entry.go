package daemon

func Setup() {
	regExitSigs()
	scheduledAutoCancel()
}
