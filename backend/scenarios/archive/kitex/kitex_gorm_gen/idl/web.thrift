namespace go user



struct ValidateWebParamRequest {
    1: optional string web_code
    2: optional string name

    // 255: optional base.Base Base
}

struct ValidateWebParamResponse {
    1: required bool isValid

    // 255: optional base.BaseResp BaseResp
}

service WebService {
    ValidateWebParamResponse ValidateWebParam(1: ValidateWebParamRequest req)
}

// TODO @huangjunqing