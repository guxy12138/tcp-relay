package org.example.util;

import org.springframework.beans.BeansException;
import org.springframework.beans.factory.support.BeanDefinitionBuilder;
import org.springframework.beans.factory.support.DefaultListableBeanFactory;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.stereotype.Component;

@Component
public class SpringContextUtil implements ApplicationContextAware {

    private static ApplicationContext applicationContext = null;

    @Override
    public void setApplicationContext(ApplicationContext applicationContext) throws BeansException {
       SpringContextUtil.applicationContext = applicationContext;
    }

    public static ApplicationContext getApplicationContext() {
        return applicationContext;
    }

    /**
     * 适用于springbean使用注解@Service("XXXService")
     * 获取接口对象 参数传入 XXXService
     *
     * @param name String
     * @return <T> bean对象
     */
    @SuppressWarnings("unchecked")
    public static <T> T getBean(String name) throws BeansException {
        return applicationContext == null ? null : (T) applicationContext.getBean(name);
    }

    /**
     * 适用于springbean使用注解@Service
     * 获取接口对象 参数传入 XXXService.class  不是 XXXServiceImpl.class
     *
     * @param name Class<T>
     * @return <T> bean对象
     */
    public static <T> T getBean(Class<T> name) throws BeansException {
        return applicationContext == null ? null : applicationContext.getBean(name);
    }

    /**
     * 动态注入bean
     *
     * @param requiredType 注入类
     * @param beanName     bean名称
     */
    public static void registerBean(Class<?> requiredType, String beanName) {

        //将applicationContext转换为ConfigurableApplicationContext
        ConfigurableApplicationContext configurableApplicationContext = (ConfigurableApplicationContext) applicationContext;

        //获取BeanFactory
        DefaultListableBeanFactory defaultListableBeanFactory = (DefaultListableBeanFactory) configurableApplicationContext.getAutowireCapableBeanFactory();

        //创建bean信息.
        BeanDefinitionBuilder beanDefinitionBuilder = BeanDefinitionBuilder.genericBeanDefinition(requiredType);

        //动态注册bean.
        defaultListableBeanFactory.registerBeanDefinition(beanName, beanDefinitionBuilder.getBeanDefinition());
    }

    /**
     * 动态销毁bean
     *
     * @param beanName bean名称
     */
    public static void destroyBean(String beanName) {
        //将applicationContext转换为ConfigurableApplicationContext
        ConfigurableApplicationContext configurableApplicationContext = (ConfigurableApplicationContext) applicationContext;
        //获取BeanFactory
        DefaultListableBeanFactory defaultListableBeanFactory = (DefaultListableBeanFactory) configurableApplicationContext.getAutowireCapableBeanFactory();
        //确定bean是否已经被注册上了
         if (defaultListableBeanFactory.isBeanNameInUse(beanName)){
             //代表已经有bean被注册了需要将其销毁
             //销毁bean
             defaultListableBeanFactory.removeBeanDefinition(beanName);
         }
    }
}

