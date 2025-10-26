package org.example.models;

import com.alibaba.fastjson2.annotation.JSONField;
import lombok.Data;
import org.example.util.BinaryTransUtils;

import java.util.List;

@Data
public class DataSegmentModel {
    //年-2Byte-16bits
    @JSONField(name = "年",ordinal = 1)
    private Integer year;
    //月-1Byte
    @JSONField(name = "月",ordinal = 2)
    private Integer month;
    //日-1Byte
    @JSONField(name = "日",ordinal = 3)
    private Integer day;
    //数据内容
    @JSONField(name = "数据内容",ordinal = 4)
    private List<DataModel> dataList;

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();
        sb.append(BinaryTransUtils.intToBinaryString(year,16));
        sb.append(BinaryTransUtils.intToBinaryString(month,8));
        sb.append(BinaryTransUtils.intToBinaryString(day,8));
        dataList.stream().forEach(dataModel -> sb.append(dataModel.toString()));
        return sb.toString();
    }
}
