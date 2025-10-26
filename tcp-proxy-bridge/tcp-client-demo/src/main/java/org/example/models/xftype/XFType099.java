package org.example.models.xftype;

import com.alibaba.fastjson2.annotation.JSONField;
import lombok.Data;
import org.example.models.MessageDataModel;

//集团客户认证结果信息
@Data
public class XFType099 extends MessageDataModel {
    //token-32Bytes
    @JSONField(name = "token")
    private String token;
    @JSONField(name = "认证状态")
    //认证状态-1Bytes
    private int certificationStatus;
}
