package org.example.handler;

import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelFuture;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInboundHandlerAdapter;
import io.netty.handler.timeout.IdleState;
import io.netty.handler.timeout.IdleStateEvent;
import lombok.extern.slf4j.Slf4j;
import org.example.models.xftype.XFType203;
import org.example.util.BinaryTransUtils;

@Slf4j
public class HeartBeatHandler extends ChannelInboundHandlerAdapter {
    @Override
    public void userEventTriggered(ChannelHandlerContext ctx, Object evt) throws Exception {
        super.userEventTriggered(ctx, evt);
        if (evt instanceof IdleStateEvent) {
            IdleStateEvent idleStateEvent = (IdleStateEvent) evt;
            //长期未发送写事件
            if (idleStateEvent.state().equals(IdleState.WRITER_IDLE)){
                log.info("3s未触发写事件,开始向服务端{}发送心跳报文",ctx.channel());
                XFType203 xfType203 = new XFType203();
                //发送心跳报文
                ChannelFuture channelFuture=ctx.channel().writeAndFlush(Unpooled.copiedBuffer(BinaryTransUtils.string2bytes(xfType203.toString())));
                if (channelFuture.isSuccess()){
                    log.info("心跳报文发送成功");
                }
            }
            if (idleStateEvent.state().equals(IdleState.READER_IDLE)){
                log.info("80s未触发读事件,开始主动断线重连");
                ctx.close();
            }
        }
    }
}
