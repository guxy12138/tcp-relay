package org.example.models.broadcast;

import lombok.Data;

@Data
public class FixedDomain {
    //卫星号-6bit
    private String satelliteNo;
    //波束号-4bit
    private String beamNo;
    //超帧计数-17bit
    private String superFrameCount;
    //周计数-13bit
    private String weekCount;
    //频度动态控制-4bit
    private String frequencyDynamicControl;
    //出站时延预报-4bit
    private String outboundDelayPrediction;
    //完好性信息-2bit
    private String integrityInfo;
    //保留1-5bit
    private String remain1;
    //保留2-9bit
    private String remain2;
    //重复（暂定）-64bit
    private String repeat;
}
