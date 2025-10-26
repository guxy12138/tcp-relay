package org.example.service;

import cn.hutool.json.JSONUtil;
import com.alibaba.fastjson2.JSONObject;
import lombok.extern.slf4j.Slf4j;
import org.example.compent.parser.XFTypeParser;
import org.example.models.*;
import org.example.models.entity.DetailInformationEntity;
import org.example.util.NumUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.stereotype.Service;

import java.nio.charset.StandardCharsets;
import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * @classname: BinaryCodeService
 * @description: 解析byte数组服务
 * @author: 王泽森
 * @date: 2023/5/17 10:22
 */
//@Service
@Component
@Slf4j
public class BinaryCodeService {
    @Autowired
    private List<XFTypeParser> parsers;

    public void basePackageParse(byte[] bytes) {
        int count = 0;
        DetailInformationEntity detailInformationEntity = new DetailInformationEntity();
        String hexString = NumUtils.bytesToHexString(bytes);
        log.info(hexString);
        if (hexString.equals("E5BF83E8B7B3") || hexString.equals("E5BF83E8B7B3".toLowerCase())) {
            log.info("收到心跳报文");
            return;
        }
        BasePackageModel basePackageModel = new BasePackageModel();
        //信源
        String xyHexString = hexString.substring(count, count += 4);
        int xyInt = Integer.parseInt(xyHexString, 16);
        log.info("信源:" + xyInt);
        basePackageModel.setSourceInfo(xyInt);

        //信宿
        String xsHexString = hexString.substring(count, count += 4);
        int xsInt = Integer.parseInt(xsHexString, 16);
        log.info("信宿:" + xsInt);
        basePackageModel.setHostInfo(xsInt);

        //包序号
        String bxhHexString = hexString.substring(count, count += 8);
        Long bxhLong = Long.parseLong(bxhHexString, 16);
        log.info("包序号:" + bxhLong);
        basePackageModel.setPackageNo(bxhLong);

        //当前数据项
        String dqsjdHexString = hexString.substring(count, count += 2);
        int dqsjdInt = Integer.parseInt(dqsjdHexString, 16);
        log.info("当前数据项:" + dqsjdInt);
        basePackageModel.setCurrentDataItem(dqsjdInt);

        //当前数据段长度
        String dqsjdzcdHexString = hexString.substring(count, count += 4);
        int dqsjdzcdInt = Integer.parseInt(dqsjdzcdHexString, 16);
        log.info("当前数据段长度:" + dqsjdzcdInt);
        basePackageModel.setDataSumLength(dqsjdzcdInt);

        //重复标志
        String cfbzHexString = hexString.substring(count, count += 2);
        int cfbzInt = Integer.parseInt(cfbzHexString, 16);
        log.info("重复标志:" + cfbzInt);
        if (cfbzInt == 0) {
            log.info("本次无重发数据");
        } else {
            log.info("本次报文中携带重发数据");
        }
        basePackageModel.setRetransmissionFlag(cfbzInt);

        //重发数据项
        String cfsjxHexString = hexString.substring(count, count += 2);
        int cfsjxInt = Integer.parseInt(cfsjxHexString, 16);
        log.info("重发数据项:" + cfsjxInt);
        basePackageModel.setRetransmissionData(cfsjxInt);

        //重发数据段长度
        String cfsjdzcdHexString = hexString.substring(count, count += 4);
        int cfsjdzcdInt = Integer.parseInt(cfsjdzcdHexString, 16);
        log.info("重发数据段长度:" + cfsjdzcdInt);
        basePackageModel.setRetransmissionSumLength(cfsjdzcdInt);

        String data = hexString.substring(count);
        //开始解析dataSegemrnt字段
        //当前数据项
        String currentDataSegmentString = data.substring(0, dqsjdzcdInt * 2);
        DataSegmentModel currentDataSegment = dataSegmentParse(currentDataSegmentString);
        DataSegmentModels dataSegmentModels = new DataSegmentModels();
        dataSegmentModels.setCurrentDataSegment(currentDataSegment);
        basePackageModel.setDataSegments(dataSegmentModels);
        log.info("数据包解析完成数据包的内容为{}", JSONObject.toJSONString(basePackageModel));
        detailInformationEntity.setSenderAddress(String.valueOf(xyInt));
        detailInformationEntity.setRecipientAddress(String.valueOf(xsInt));
        String deatail = JSONUtil.toJsonStr(basePackageModel);

        detailInformationEntity.setDetails(deatail);
//        detailInformationService.save(detailInformationEntity);
    }

    public DataSegmentModel dataSegmentParse(String currentDataSegmentString) {
        log.info("开始解析数据段");
        DataSegmentModel dataSegmentModel = new DataSegmentModel();
        int count = 0;
        //年
        String yearString = currentDataSegmentString.substring(count, count += 4);
        int year = Integer.parseInt(yearString, 16);
        dataSegmentModel.setYear(year);
        log.info("年:{}", year);
        //月
        String monthString = currentDataSegmentString.substring(count, count += 2);
        int month = Integer.parseInt(monthString, 16);
        dataSegmentModel.setMonth(month);
        log.info("月:{}", month);
        //日
        String dayString = currentDataSegmentString.substring(count, count += 2);
        int day = Integer.parseInt(dayString, 16);
        dataSegmentModel.setDay(day);
        log.info("日:{}", day);
        //数据内容
        String dataModelString = currentDataSegmentString.substring(count, currentDataSegmentString.length());
        List<DataModel> dataModels = new ArrayList<>();
        dataModelParse(dataModels, dataModelString);
        dataSegmentModel.setDataList(dataModels);
        return dataSegmentModel;
    }

    public void dataModelParse(List<DataModel> dataModels, String dataModelString) {
        log.info("开始解析服务信息内容");
        DataModel dataModel = new DataModel();
        int count = 0;
        //信息类型编号
        String messageTypeNoString = dataModelString.substring(count, count += 4);
        int messageTypeNo = Integer.parseInt(messageTypeNoString, 16);
        dataModel.setMessageTypeNo(messageTypeNo);
        log.info("信息类型编号:{}", messageTypeNo);
        //信息内容长度
        String messageLengthString = dataModelString.substring(count, count += 4);
        int messageLength = Integer.parseInt(messageLengthString, 16);
        dataModel.setMessageLength(messageLength);
        log.info("信息内容长度:{}", messageLength);
        //信息内容
        //当前信息段的内容
        String message_data = dataModelString.substring(count, count + messageLength * 2);
        List<MessageDataModel> messageDataModelList = parsers.stream().filter(xfTypeParser -> xfTypeParser.support(messageTypeNo)).map(xfTypeParser -> {
            try {
                return xfTypeParser.parse(message_data);
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }).collect(Collectors.toList());
        if (messageDataModelList.size() > 0 && messageDataModelList.get(0) != null) {
            MessageDataModel messageDataModel = messageDataModelList.get(0);
            Optional.ofNullable(messageDataModel).ifPresent(dataModel::setMessageData);
            dataModels.add(dataModel);
            if (count + messageLength * 2 == dataModelString.length()) {
                return;
            }
            //余下信息段的内容
            String remain_data = dataModelString.substring(count + messageLength * 2, dataModelString.length());
            dataModelParse(dataModels, remain_data);
        } else {
            log.warn("还没解！！！！！！！！！");
            log.warn("余下数据:" + message_data);
        }

    }
}
