package org.example.util.message.xftype100;

import org.example.models.BasePackageModel;
import org.example.models.DataModel;
import org.example.models.DataSegmentModel;
import org.example.models.DataSegmentModels;
import org.example.models.xftype.XFType100;
import org.example.util.BinaryTransUtils;
import org.example.util.NumUtils;
import org.example.util.NumberUtil;
import org.example.util.message.BasePackageInit;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.nio.charset.StandardCharsets;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;

public class BaseInitXFtype100 {
    private final Random rand = SecureRandom.getInstanceStrong();

    public BaseInitXFtype100() throws NoSuchAlgorithmException {
        ;
    }

    public static BasePackageModel initXFType100() throws NoSuchAlgorithmException {
//        String tokenString = "e9b00329a75f4b0f646ad98e0d4a2771";
        String tokenString = "11111111111111111111111111111111";
//        String token = "3131313131313131313131313131313131313131313131313131313131313131";
        String token = NumUtils.bytesToHexString(tokenString.getBytes(StandardCharsets.US_ASCII));

        XFType100 xfType100 = new XFType100();
        xfType100.setToken(BinaryTransUtils.hexStrToBinaryStr(token));
        //所有信息段
        List<DataModel> dataModels = new ArrayList<DataModel>();
        //初始化当前信息段
        DataModel dataModel = BasePackageInit.iniDataModel(100, xfType100);
        //后面共32*8个位的二进制字符串
        dataModel.setMessageLength(32);
        //加入信息段中
        dataModels.add(dataModel);
        //初始化当前数据段
        DataSegmentModel dataSegment = BasePackageInit.initDataSegment(dataModels);
        //初始化无重发数据段的数据段
        DataSegmentModels dataSegmentModels = BasePackageInit.initDataSegmentModels(dataSegment, null);
        //初始化BasePackage
        BasePackageModel basePackage = BasePackageInit.initBasePackageModel(1L, dataSegmentModels);
        //设置PackageNo
        basePackage.setPackageNo(1L);
        //设置当前数据总长度
        basePackage.setDataSumLength(0x0028);
        //信宿
        basePackage.setHostInfo(0x0014);
        //信源
//        //990
//        basePackage.setSourceInfo(0x3DE);
        //802
        basePackage.setSourceInfo(0x322);
        //当前数据项
        basePackage.setCurrentDataItem(0x01);
        //重发标志
        basePackage.setRetransmissionFlag(0x00);
        //重发数据项
        basePackage.setRetransmissionData(0x00);
        //重发数据总长度
        basePackage.setRetransmissionSumLength(0x0000);
        return basePackage;
    }
}
