// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package params

// MainnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the main Abeychain network.
var MainnetBootnodes = []string{
	"enode://031569aae8f6ab0bb19fc0feb719b5312fffcee3f703e8719aeca6cc1ae338889e23f85d6d4f92926dbf62aa2c8d7f40637e9202b34e0cfdb6910184cc5f0a39@3.66.27.8:30313",
	"enode://49c2cd85519592bb9094007ae55b46588ca478c3208cad64c33af8d538ad2537cdc7951063ca3b5f6d6f4d4b4247b98e037100b1c9f9e9dfd07f1b113496d324@52.74.197.57:30313",
	"enode://3ca215e57c110fc516f7c6430fe61933c4791d0a962958e55c0b3d0066695c8f69f6f2cf5f41ed1ae4887a0f35f6e6c1774c916b8f7ea95c2f70c0548c0c9a03@52.58.40.204:30313",
	"enode://a9a9bd0c9b9c7aa127cd14f249f15ff476276c0e0840269b8fcdb1a593e83cd6893a6151a681f2621e4dcb522394e715e8c2a906a7cae4e751d4e374c2216702@52.220.223.77:30313",
	"enode://1fffc81ce673cdc5e4ebc43888f4ee3e7a1fce944b152289d435b6b818ecc5189fc0695f53f5330ef9435785501cfc5608a6246d429078a2af62886eec57566b@54.227.239.172:30313",
	"enode://d8ba14421c73f89f51e00bffb76429eef9f0df18422f3b5b49f448be52cdcf0b375c510a63cf019c53dd40dc3466525cc724e202b4e512c670990d3c722074a8@13.52.29.44:30313",
}

// TestnetBootnodes are the enode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"enode://cbbd20d50bda36ac0a52ca4d48441cb5b3438686ac0faa13f69818d7498a52e227ee4975fecafb4828458e423d1cfee4aadbd442a889213e16029e18063e7d25@10.0.196.190:30313",
	"enode://1b4fd85fe9e7de46e60e7893447d5fd7ce3f3c613c2c4b54bee2391a0a9b4de2a08f949fbbec01c6a4881f2ff37edd60b00426fad4208d925bcaafacbcc342d2@10.0.196.12:30313",
	"enode://8f5c4da356913c0210a3b9325a2e931ecb2ae6f9f898e813dd3c0da6dcb88ad0c5117c3c0ec40d73d214dca32e0e2a89844a51895f2c9b8632b558a71a57376c@10.0.196.26:30313",
	"enode://1cf7284d4933f670a9c53cc28b1d1e7da4edc693e9b77b9a4a4bafaa17a121ed8b4b61b91575a3abce1d9b11bbe3a9da8534c89ccefffc2d3936fda29da0df3c@10.0.196.216:30313",
}

// DevnetBootnodes are the enode URLs of the P2P bootstrap nodes running on
// the dev Abeychain network.
var DevnetBootnodes = []string{
	//"enode://e8e1af516589b5b49d0110bf52dcbc8c4fe592697cfd94eb01fdc302e2b4f1f9c7d3667ca4be64d094a15635204195681219dc59febe17c7c83f3b71b6258454@18.138.171.105:30313",
	//"enode://8b1af471dd393f2f228cefd663e3176963592adcd452e69b868ffb2cbb2f7b67ca44f60269e031952a00ba708799c8c1c2a41c279df80461e9b4d384ea49f9fc@18.140.45.222:30313",
	//"enode://1554bd2e50456dbf1e5b94aabf88fc7eb1f274d39ae52067541d46e2b8043cb102bffffecb0ba42657c8904aec6bde4f1728110107f27f50616c18ede4af983c@52.220.163.103:30313",
	//"enode://d3b5fb4283424e6011d6ad1bcad7e3890fc94db4e6d221571a61985b1f48b6ed26733b9871debb18924cb299600611b683f08e1be08e9a320ffba44494388d1f@54.151.132.19:30313",

	"enode://ea909005a7e2caccfcd6bb460c03bc7a0c700810a6e4656b7a4baa108f6d2935d40e88d68b89cc1d23dbbbd5800a77530fdf4af39006250db1baf426bd6b915e@127.0.0.1:30501",
	"enode://57eaa41e0d4fa10ac34f4f82526a53044535494087c234b3284d5ba664c3d4c586f752331b735092c97ca2d91b9a5fa05cdf4c079c044200240cf882186396f1@127.0.0.1:30601",
	"enode://06232b359e51ee21d4d60c546d557cdf4140d5ac15392d8a6bce0fa47875773b263152a1597d690c0f9faf836d666c2b28a9916089968eee4498a2cd6cb7088c@127.0.0.1:30701",
	"enode://7c822377b4cd6989b3285a5ce35bfa7fd0c2c69924c2989545247e855e47dfe76ea3480d9723e72534fb91b7d8157419cf05196f1183db06fe299ebe099774a5@127.0.0.1:30801",
}

// DiscoveryV5Bootnodes are the enode URLs of the P2P bootstrap nodes for the
// experimental RLPx v5 topic-discovery network.
var DiscoveryV5Bootnodes = []string{
	"enode://ebb007b1efeea668d888157df36cf8fe49aa3f6fd63a0a67c45e4745dc081feea031f49de87fa8524ca29343a21a249d5f656e6daeda55cbe5800d973b75e061@39.98.171.41:30315",
	"enode://b5062c25dc78f8d2a8a216cebd23658f170a8f6595df16a63adfabbbc76b81b849569145a2629a65fe50bfd034e38821880f93697648991ba786021cb65fb2ec@39.98.43.179:30312",
}
