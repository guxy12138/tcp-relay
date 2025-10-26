package org.example.models.broadcast;

import lombok.Data;

import java.util.List;

@Data
public class SatelliteInfo {
    //导航PRN号
    private String navigationNo;
    //波束号-4bit
    private String beamNo;
    //保留1
    private String remain1;
    //保留2
    private String remain2;
    //波束中心经纬度
    List<Position> positions;
}
