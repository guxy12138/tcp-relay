package org.example.service;

import io.netty.buffer.Unpooled;
import io.netty.channel.Channel;
import lombok.extern.slf4j.Slf4j;
import org.example.models.BasePackageModel;
import org.example.models.RespBean;
import org.example.util.BinaryTransUtils;
import org.example.util.SpringContextUtil;
import org.springframework.stereotype.Service;

import java.util.Optional;

/**
 * @author wangzesen
 */
@Service
@Slf4j
public class NettyClientService {
    //入参为需要转发的包结构
    public RespBean<Object> sendMessage(BasePackageModel basePackageModel){
        Channel channel = SpringContextUtil.getBean("clientChannel");
        String msg = basePackageModel.toString();
        if (Optional.ofNullable(channel).isPresent()){
            byte[] bytes = BinaryTransUtils.string2bytes(msg);
            channel.writeAndFlush(Unpooled.copiedBuffer(bytes));
        }
        return RespBean.ok();
    }
}
