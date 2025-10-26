package org.example.models;

import java.io.Serializable;

public class RespBean<T> implements Serializable {

    private static final long serialVersionUID = 1L;

    private Integer code;
    private String msg;
    private T result;

    public RespBean() {
    }

    public RespBean(Integer code, String msg, T result) {
        this.code = code;
        this.msg = msg;
        this.result = result;
    }

    public static <T> RespBean<T> build() {
        return new RespBean<>();
    }

    public static <T> RespBean<T> ok() {
        return new RespBean<>(200, "请求成功", null);
    }

    public static <T> RespBean<T> ok(T result) {
        return new RespBean<>(200, "请求成功", result);
    }

    public static <T> RespBean<T> ok(String msg, T result) {
        return new RespBean<>(200, msg, result);
    }

    public static <T> RespBean<T> msg(String msg) {
        return new RespBean<>(200, msg, null);
    }

    public static <T> RespBean<T> error() {
        return new RespBean<>(500, "请求失败", null);
    }

    public static <T> RespBean<T> error(String msg) {
        return new RespBean<>(500, msg, null);
    }

    public static <T> RespBean<T> error(int code, String msg) {
        return new RespBean<>(code, msg, null);
    }

    public Integer getCode() {
        return code;
    }

    public void setCode(Integer code) {
        this.code = code;
    }

    public String getMsg() {
        return msg;
    }

    public void setMsg(String msg) {
        this.msg = msg;
    }

    public T getResult() {
        return result;
    }

    public void setResult(T result) {
        this.result = result;
    }
}