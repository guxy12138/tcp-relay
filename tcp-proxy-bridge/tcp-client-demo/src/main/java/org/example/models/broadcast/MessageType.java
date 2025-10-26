package org.example.models.broadcast;

import lombok.Data;

import java.util.List;

@Data
public class MessageType {
    //本帧电文包括卫星数
    private String mesSatellites;
    //卫星信息
    private List<SatelliteInfo> satelliteInfos;
    //保留字段
    private String reamin;
}