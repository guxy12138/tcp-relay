package org.example.xftype001;

import org.example.models.xftype.XFType001;
import org.springframework.context.annotation.Bean;
import org.springframework.stereotype.Component;

public class BaseInitXFType001 {
    public XFType001 initXFType001() {
        XFType001 xfType001 = new XFType001();
        //固定为1，代表区域服务
        xfType001.setDeparture("00000001");
        //固定为0，表示收方地址低位24bit有效
        xfType001.setReceiverType("00000000");
        //报告模式"010"RNSS位置报告结果
        xfType001.setReportModel("010");
        //信息类别字段
        //保留位
        xfType001.setRemain(0);
        //出站帧类别
        xfType001.setStackOutType("00");
        return xfType001;
    }
}