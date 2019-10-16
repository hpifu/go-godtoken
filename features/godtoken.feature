Feature: GetToken

    Scenario: get token
        Given redis set string "godtoken"
            """
            d571bda90c2d4e32a793b8a1ff4ff984
            """
        When grpc 请求 godtoken GetToken
            """
            {}
            """
        Then grpc 检查
            """
            {
                "token": "d571bda90c2d4e32a793b8a1ff4ff984"
            }
            """
        Given redis del "godtoken"
