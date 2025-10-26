package org.example.compent.parser;

import lombok.extern.slf4j.Slf4j;
import org.example.models.MessageDataModel;
import org.example.models.xftype.XFType099;
import org.example.util.NumUtils;
import org.springframework.stereotype.Component;

import java.nio.charset.StandardCharsets;

@Component
@Slf4j
public class XFType99Parser implements XFTypeParser {
    @Override
    public MessageDataModel parse(String data) throws Exception {
        log.info("开始解析XFType099的数据");
        XFType099 xfType099 = new XFType099();
        int count = 0;
        String tokenString = data.substring(0,count+=64);
        String token = new String(NumUtils.hexStringToBytes(tokenString), StandardCharsets.US_ASCII);
        log.info("token为{}",token);
        String statesString = data.substring(count,count+=2);
        int state = Integer.parseInt(statesString, 16);
        log.info("认证状态为{}",state);
        if(state == 0){
            log.error("认证失败");
        }
        if (state == 1){
            log.info("认证成功");
        }
        xfType099.setToken(token);
        xfType099.setCertificationStatus(state);
        return xfType099;
    }

    @Override
    public Boolean support(int messageTypeNo) {
        return messageTypeNo == 99;
    }
}
