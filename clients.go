package yt_transcript

// clientIdentity is a YouTube innertube client. Fields must be mutually
// consistent: the user agent, client version, and OS/device fields are all
// tied together, and YouTube rejects mismatched combinations.
type clientIdentity struct {
	name        string
	version     string
	androidSDK  int
	deviceMake  string
	deviceModel string
	osName      string
	osVersion   string
	userAgent   string
}

// context builds the innertube request context for this client.
func (ci clientIdentity) context() map[string]any {
	client := map[string]any{
		"clientName":    ci.name,
		"clientVersion": ci.version,
	}
	if ci.androidSDK > 0 {
		client["androidSdkVersion"] = ci.androidSDK
	}
	if ci.deviceMake != "" {
		client["deviceMake"] = ci.deviceMake
	}
	if ci.deviceModel != "" {
		client["deviceModel"] = ci.deviceModel
	}
	if ci.osName != "" {
		client["osName"] = ci.osName
	}
	if ci.osVersion != "" {
		client["osVersion"] = ci.osVersion
	}
	return client
}

// clientIdentities are tried in turn; the web client is absent because it is
// poToken-gated. Rotating across families, not versions, is what adds
// fingerprint diversity. All entries are verified by TestLive_FetchEachClient.
var clientIdentities = []clientIdentity{
	{
		name: "ANDROID", version: "21.03.36", androidSDK: 36,
		deviceMake: "samsung", deviceModel: "SM-S908E", osName: "Android", osVersion: "16",
		userAgent: "com.google.android.youtube/21.03.36 (Linux; U; Android 16; en_US; SM-S908E Build/TP1A.220624.014) gzip",
	},
	{
		name: "iOS", version: "20.11.6",
		deviceMake: "Apple", deviceModel: "iPhone10,4", osName: "iOS", osVersion: "16.7.7.20H330",
		userAgent: "com.google.ios.youtube/20.11.6 (iPhone10,4; U; CPU iOS 16_7_7 like Mac OS X)",
	},
	{
		name: "ANDROID_VR", version: "1.65.10", androidSDK: 32,
		deviceMake: "Oculus", deviceModel: "Quest 3", osName: "Android", osVersion: "12L",
		userAgent: "com.google.android.apps.youtube.vr.oculus/1.65.10 (Linux; U; Android 12L; eureka-user Build/SQ3A.220605.009.A1) gzip",
	},
	{
		name: "VISIONOS", version: "1.02",
		deviceMake: "Apple", deviceModel: "RealityDevice17,1", osName: "visionOS", osVersion: "26.5.23O471",
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 15_7_3) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.0 Safari/605.1.15",
	},
}
