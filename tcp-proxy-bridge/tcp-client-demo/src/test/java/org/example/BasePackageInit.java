package org.example;

import org.example.models.*;
import org.example.util.BinaryTransUtils;

import java.util.Calendar;
import java.util.Date;
import java.util.List;

public class BasePackageInit {
    //初始化包的报文需要传入package号
    public BasePackageModel initBasePackageModel(String packageNo, DataSegmentModels dataSegment) {
        BasePackageModel basePackage = new BasePackageModel();
        //配置信源
        basePackage.setSourceInfo(0x0014);
        //配置信宿
        basePackage.setHostInfo(0x0322);
        //包序号
        basePackage.setPackageNo(0L);
        //当前数据项
        basePackage.setCurrentDataItem(0x01);
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
    public DataSegmentModels initDataSegmentModels(DataSegmentModel currentDataSegment, DataSegmentModel retransmissionDataSegment) {
        DataSegmentModels dataSegments = new DataSegmentModels();
        dataSegments.setCurrentDataSegment(currentDataSegment);
        dataSegments.setRetransmissionDataSegment(retransmissionDataSegment);
        return dataSegments;
    }

    //初始化数据段(包括当前数据段,重发数据段)
    public DataSegmentModel initDataSegment(List<DataModel> dataModels) {
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

    public DataModel iniDataModel(String messageTypeNo, MessageDataModel messageData) {
        DataModel dataModel = new DataModel();
        dataModel.setMessageTypeNo(0);
        dataModel.setMessageData(messageData);
        dataModel.setMessageLength(dataModel.getMessageData().sumLength());
        return dataModel;
    }
}
