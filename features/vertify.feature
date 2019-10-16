Feature: verify

    Scenario: verify failed
        Given redis set string "godtoken"
            """
            d571bda90c2d4e32a793b8a1ff4ff984
            """
        When grpc 请求 godtoken Verify
            """
            {
                "token": "123"
            }
            """
        Then grpc 检查
            """
            {
                "ok": false
            }
            """
        Given redis del "godtoken"

    Scenario: verify success
        Given redis set string "godtoken"
            """
            d571bda90c2d4e32a793b8a1ff4ff984
            """
        When grpc 请求 godtoken Verify
            """
            {
                "token": "d571bda90c2d4e32a793b8a1ff4ff984"
            }
            """
        Then grpc 检查
            """
            {
                "ok": true
            }
            """
        Given redis del "godtoken"
