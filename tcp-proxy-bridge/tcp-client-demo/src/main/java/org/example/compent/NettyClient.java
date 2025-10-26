package org.example.compent;

import io.netty.bootstrap.Bootstrap;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.channel.*;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.DelimiterBasedFrameDecoder;
import io.netty.handler.logging.LogLevel;
import io.netty.handler.logging.LoggingHandler;
import io.netty.handler.ssl.SslContext;
import io.netty.handler.ssl.SslContextBuilder;
import io.netty.handler.ssl.util.SelfSignedCertificate;
import io.netty.handler.timeout.IdleStateHandler;
import lombok.extern.slf4j.Slf4j;
import org.example.config.TcpClientConfiguration;
import org.example.exception.CustomExceptionHandler;
import org.example.handler.HeartBeatHandler;
import org.example.service.BinaryCodeService;
import org.example.util.BinaryTransUtils;
import org.example.util.message.xftype100.BaseInitXFtype100;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.DependsOn;
import org.springframework.stereotype.Component;

import javax.annotation.PostConstruct;
import javax.net.ssl.SSLException;
import java.io.File;
import java.security.spec.PKCS8EncodedKeySpec;
import java.util.Optional;
import java.util.concurrent.TimeUnit;

@Component
@DependsOn("springContextUtil")
@Slf4j
public class NettyClient {

    private EventLoopGroup eventLoop = new NioEventLoopGroup();
    @Autowired
    private BinaryCodeService binaryCodeService;
    @Autowired
    private TcpClientConfiguration tcpClientConfiguration;


    /**
     * netty client 连接，连接失败10秒后重试连接
     */
    @PostConstruct
    public void initConnect(){
        log.info("开始初始化连接......");
        connect(new Bootstrap());
    }

    public void connect(Bootstrap bootstrap) {
        try {
//            Optional.ofNullable(bootstrap).ifPresent(this::initChannel);
            initChannel(bootstrap);
        } catch (Exception e) {
            log.error("连接客户端失败,error：" + e);
        }
    }

    private void initChannel(Bootstrap bootstrap) throws SSLException {

        bootstrap.group(eventLoop);
        bootstrap.channel(NioSocketChannel.class);
        bootstrap.option(ChannelOption.SO_KEEPALIVE, true);
//        //引入SSL安全验证
//        File certChainFile = new File("E:\\workspace\\zs\\990\\cert.crt");
//        File keyFile = new File("E:\\workspace\\zs\\990\\pkcs8.pem");
//        File rootFile = new File("E:\\workspace\\zs\\990\\ca.crt");

//        // netty sslContextBuilder 使用客户端cert+key.pk8，ca
//        SslContext sslCtx = SslContextBuilder.forClient()
//                //客户端 crt+key.pk8
//                .keyManager(certChainFile, keyFile)
//                //ca公钥
//                .trustManager(rootFile)
//                .build();

        bootstrap.handler(new ChannelInitializer<SocketChannel>() {
            @Override
            protected void initChannel(SocketChannel socketChannel) {

                //配置TLS/SSL安装验证
//                socketChannel.pipeline().addLast(sslCtx.newHandler(socketChannel.alloc()));

                //分割符
                String separatorCharacter = BinaryTransUtils.hexStrToBinaryStr("7878787888888888");
                //将分割符转化为二进制码
                byte[] bytes = BinaryTransUtils.string2bytes(separatorCharacter);
                ByteBuf byteBuf = Unpooled.copiedBuffer(bytes);
                //配置定时发送心跳检测，触发器
                //1.读操作空闲时间 2.写操作空闲时间 3.读写操作空闲时间
                socketChannel.pipeline().addLast("idleState", new IdleStateHandler(60, 3, 60 * 10, TimeUnit.SECONDS));
                socketChannel.pipeline().addLast(new DelimiterBasedFrameDecoder(4096, true, true, byteBuf));
                //配置客户端处理器
                socketChannel.pipeline().addLast(new ChannelInboundHandlerAdapter() {
                    @Override
                    public void channelRead(ChannelHandlerContext ctx, Object msg) throws Exception {
                        //读事件触发
                        ByteBuf byteBuf = (ByteBuf) msg;
                        int size = byteBuf.readableBytes();
                        log.info("收到客户端发来的信息长度为{}", byteBuf);
                        byte[] bytes = new byte[size];
                        byteBuf.readBytes(bytes);
                        BinaryCodeService binaryParser = Optional.ofNullable(binaryCodeService).orElseThrow(() -> new Exception("未找到binaryCodeService"));
                        try {
                            binaryParser.basePackageParse(bytes);
                        } catch (Exception e) {
                            e.printStackTrace();
                        }
                        super.channelRead(ctx, msg);
                    }

                    @Override
                    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
                        log.error("与客户端的链接断开,channel的信息为{}", ctx.channel());
                        //断线机制触发，需要重连
                        connect(new Bootstrap());
                    }

                    /**
                     * 当连接建立成功后，触发的代码逻辑。
                     * 在一次连接中只运行唯一一次。
                     * 通常用于实现连接确认和资源初始化的。
                     */
                    @Override
                    public void channelActive(ChannelHandlerContext ctx) throws Exception {
                        log.debug("与服务端的链接成功,channel的信息为{}", ctx.channel());
                        log.info("开始发送身份认证信息");
                        String msg = BaseInitXFtype100.initXFType100().toString();
                        byte[] bytes = BinaryTransUtils.string2bytes(msg);
                        //发送身份认证信息
                        ctx.channel().writeAndFlush(Unpooled.copiedBuffer(bytes));
                    }
                });
                //配置日志
                socketChannel.pipeline().addLast(new LoggingHandler(LogLevel.INFO));
                //配置心跳处理机制
                socketChannel.pipeline().addLast(new HeartBeatHandler());
                //添加异常处理机制
                socketChannel.pipeline().addLast(new CustomExceptionHandler());
            }
        });
        bootstrap.remoteAddress(tcpClientConfiguration.getAddress(), tcpClientConfiguration.getPort());
        bootstrap.connect().addListener((ChannelFuture futureListener) -> {
            final EventLoop eventLoop = futureListener.channel().eventLoop();
            if (!futureListener.isSuccess()) {
                log.warn("客户端已启动，与服务端建立连接失败,10s之后尝试重连!");
                eventLoop.schedule(() -> connect(new Bootstrap()), 10, TimeUnit.SECONDS);
            }
        });
    }
}
