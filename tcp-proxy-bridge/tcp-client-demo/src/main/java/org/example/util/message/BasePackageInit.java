package org.example.util.message;

import lombok.extern.slf4j.Slf4j;
import org.example.models.*;
import org.example.util.BinaryTransUtils;
import org.springframework.stereotype.Component;

import java.util.Calendar;
import java.util.Date;
import java.util.List;
//初始化Package
public class BasePackageInit {
    //初始化包的报文需要传入package号
    public static BasePackageModel initBasePackageModel(Long packageNo, DataSegmentModels dataSegment) {
        BasePackageModel basePackage = new BasePackageModel();


        //配置信源
        basePackage.setSourceInfo(0x3DE);
        //配置信宿
        basePackage.setHostInfo(0x0014);
        //包序号
        basePackage.setPackageNo(packageNo);
        //当前数据项
        basePackage.setCurrentDataItem(0);
        //当前数据总长度
        basePackage.setDataSumLength(1);
        //重发标志
        basePackage.setRetransmissionFlag(0);
        //重发数据项
        basePackage.setRetransmissionData(0);
        //重发数据总长度
        basePackage.setRetransmissionSumLength(0);
        //数据段
        basePackage.setDataSegments(dataSegment);
        return basePackage;
    }

    //初始化信息段
    public static DataSegmentModels initDataSegmentModels(DataSegmentModel currentDataSegment, DataSegmentModel retransmissionDataSegment) {
        DataSegmentModels dataSegments = new DataSegmentModels();
        dataSegments.setCurrentDataSegment(currentDataSegment);
        dataSegments.setRetransmissionDataSegment(retransmissionDataSegment);
        return dataSegments;
    }

    //初始化数据段(包括当前数据段,重发数据段)
    public static DataSegmentModel initDataSegment(List<DataModel> dataModels) {
        DataSegmentModel dataSegment = new DataSegmentModel();
        Calendar cal = Calendar.getInstance();
        Date date = new Date();
        //现在的日期
        cal.setTime(date);
        //获取年
        Integer year = cal.get(Calendar.YEAR);
        //获取月（月份从0开始，如果按照中国的习惯，需要加一）
        Integer month = cal.get(Calendar.MONTH) + 1;
        //获取日（月中的某一天）
        Integer day_month = cal.get(Calendar.DAY_OF_MONTH);
        dataSegment.setYear(year);
        dataSegment.setMonth(month);
        dataSegment.setDay(day_month);
        dataSegment.setDataList(dataModels);
        return dataSegment;
    }

    public static DataModel iniDataModel(Integer messageTypeNo, MessageDataModel messageData) {
        DataModel dataModel = new DataModel();
        dataModel.setMessageTypeNo(Integer.valueOf(messageTypeNo));
        dataModel.setMessageData(messageData);
        dataModel.setMessageLength(messageData.sumLength()/8);//长度为字节数
        return dataModel;
    }
}
