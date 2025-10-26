package org.example.post;

import org.example.Application;
import org.example.BasePackageInit;
import org.example.models.BasePackageModel;
import org.example.models.DataModel;
import org.example.models.DataSegmentModel;
import org.example.models.DataSegmentModels;
import org.example.models.message.GroundNetworkMessageSegment;
import org.example.models.xftype.XFType001;
import org.example.util.BinaryTransUtils;
import org.example.util.NumberUtil;
import org.example.xftype001.BaseInitXFType001;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import java.io.*;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.net.SocketAddress;
import java.security.NoSuchAlgorithmException;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.IntStream;

@RunWith(SpringRunner.class)
@SpringBootTest(classes = Application.class)
//发送应急搜救请求服务信息测试类
public class PostXFType001Test {
    private final BaseInitXFType001 baseInitXFType001 = new BaseInitXFType001();
    private final BasePackageInit basePackageInit = new BasePackageInit();

    //呼救申请类报文
    @Test
    public void rescueApplicationMessage() throws IOException, NoSuchAlgorithmException {
        XFType001 xfType001Message = baseInitXFType001.initXFType001();
        //发送方地址
        xfType001Message.setDeparture(NumberUtil.randomGenerationByteString(24));
        //发方入站响应卫星
        xfType001Message.setSenderResSatellite("00001001");
        //发方入站波束号
        xfType001Message.setSenderStackBeam("00001001");
        //有效位
        StringBuilder significantBit = new StringBuilder();
        //无效位
        StringBuilder invalidBit = new StringBuilder();
        //随机生成24为有效位，无效位开始补0
        IntStream.range(1, 24).forEach(value -> {
            int randomInt = (int) (Math.random() * 1);
            //生成有效位
            significantBit.append(Integer.toBinaryString(randomInt));
            //拼无效位
            invalidBit.append(Integer.toBinaryString(0));
        });
        //收方地址
        xfType001Message.setReceiverAddress(significantBit.toString() + invalidBit.toString());
        //紧急标识(随机生成)
        xfType001Message.setEmergencyFlag("0" + Integer.toBinaryString((int) (Math.random() * 1)));
        //地面网出站信息段
        GroundNetworkMessageSegment groundNetworkMessageSegment = new GroundNetworkMessageSegment();
        //本次出站结果的报告入站地址
        groundNetworkMessageSegment.setSenderAddress(NumberUtil.randomGenerationByteString(24));
        //位置报告数据
        groundNetworkMessageSegment.setLocationReportData(NumberUtil.randomGenerationByteString(93));
        //搜救业务类型-呼救申请
        groundNetworkMessageSegment.setSearchRescueType("0000");
        //搜救业务数据
        groundNetworkMessageSegment.setSearchRescueData(NumberUtil.randomGenerationByteString((int) (Math.random() * 468)));
        //信息长度
        xfType001Message.setMessageLength(8 + groundNetworkMessageSegment.toString().length());
        //所有信息段
        List<DataModel> dataModels = new ArrayList<DataModel>();
        //初始化当前信息段
        DataModel dataModel = basePackageInit.iniDataModel(BinaryTransUtils.intToBinaryString(1, 16), xfType001Message);
        //加入信息段中
        dataModels.add(dataModel);
        //初始化当前数据段
        DataSegmentModel dataSegment = basePackageInit.initDataSegment(dataModels);
        //初始化无重发数据段的数据段
        DataSegmentModels dataSegmentModels = basePackageInit.initDataSegmentModels(dataSegment, null);
        //初始化BasePackage
        BasePackageModel basePackage = basePackageInit.initBasePackageModel(NumberUtil.randomGenerationByteString(32),dataSegmentModels);
        basePackage.setDataSumLength(basePackage.getDataSegments().toString().length());
        //链接客户端，建立tcp连接
        //连接服务器
        Socket tcpClientSocket = new Socket ();//起一个客户端端口
        SocketAddress serverSocketAddress = new InetSocketAddress("192.168.127.7",8080);//服务器ip+端口
        tcpClientSocket.connect (serverSocketAddress);//tcp是面向连接的
        try{
            //通过字节流直接写入信息
            OutputStream os = tcpClientSocket.getOutputStream ();
            os.write(basePackage.toString().getBytes());
            os.close();
//            //通过字节流，直接读取数据
//            InputStream is = tcpClientSocket.getInputStream ();//获取此端口的输入流，即服务器回复的消息
//            BufferedReader reader = new BufferedReader (new InputStreamReader(is,"UTF-8"));
//            String response = reader.readLine ();
//            System.out.println ("收到回复：" + response);
        }catch (Exception e){
            e.printStackTrace();
        }

    }
}
