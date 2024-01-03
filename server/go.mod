module donnan/LoveSpeak/Server

require(
    donnan/LSpeak/lib v1.2.3
)

replace (
    donnan/LSpeak/lib => ../lib
)

go 1.21.5
