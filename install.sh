#!/bin/bash

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
plain='\033[0m'
cur_dir=$(pwd)

# check root
[[ $EUID -ne 0 ]] && echo -e "${red}错误: ${plain} 请以root权限运行此脚本 \n " && exit 1

check_arch(){
    arch=$(arch)
    if [[ $arch == "x86_64" || $arch == "x64" || $arch == "amd64" ]]; then
        arch="amd64"
    elif [[ $arch == i*86 || $arch == "x86" ]]; then
        arch="386"
    elif [[ $arch == "aarch64" || $arch == "arm64" || $arch == armv8* ]]; then
        arch="arm64"
    elif [[ $arch == "armv7l" || $arch == "armv7" || $arch == arm* ]]; then
        arch="armv7"
    elif [[ $arch == "armv6l" || $arch == "armv6" ]]; then
        arch="armv6"
    elif [[ $arch == "armv5l" || $arch == "armv5" ]]; then
        arch="armv5"
    elif [[ $arch == "s390x" ]]; then
        arch="s390x"
    else
        echo -e "${red}检测到您的架构不支持，请联系作者！${plain}"
        exit 1
    fi

    echo "架构: ${arch}"
}

check_release(){
    if [[ -e /etc/os-release ]]; then
        . /etc/os-release
        release=$ID
    elif [[ -e /usr/lib/os-release ]]; then
        . /usr/lib/os-release
        release=$ID
    fi
    os_version=$(echo $VERSION_ID | cut -d \" -f2 | cut -d . -f1)

    if [[ "${release}" == "arch" ]]; then
        echo "您的系统是 Arch Linux"
    elif [[ "${release}" == "parch" ]]; then
        echo "您的系统是 Parch linux"
    elif [[ "${release}" == "manjaro" ]]; then
        echo "您的系统是 Manjaro"
    elif [[ "${release}" == "armbian" ]]; then
        echo "您的系统是 Armbian"
    elif [[ "${release}" == "opensuse-tumbleweed" ]]; then
        echo "您的系统是 OpenSUSE Tumbleweed"
    elif [[ "${release}" == "centos" ]]; then
        if [[ ${os_version} -lt 9 ]]; then
            echo -e "${red} 请使用CentOS 9或以上版本!${plain}\n" && exit 1
        fi
    elif [[ "${release}" == "ubuntu" ]]; then
        if [[ ${os_version} -lt 22 ]]; then
            echo -e "${red} 请使用Ubuntu 22或以上版本!${plain}\n" && exit 1
        fi
    elif [[ "${release}" == "fedora" ]]; then
        if [[ ${os_version} -lt 36 ]]; then
            echo -e "${red} 请使用Fedora 36或以上版本!${plain}\n" && exit 1
        fi
    elif [[ "${release}" == "debian" ]]; then
        if [[ ${os_version} -lt 12 ]]; then
            echo -e "${red} 请使用Debian 12或以上版本!${plain}\n" && exit 1
        fi
    elif [[ "${release}" == "almalinux" ]]; then
        if [[ ${os_version} -lt 95 ]]; then
            echo -e "${red} 请使用AlmaLinux 9.5或以上版本!${plain}\n" && exit 1
        fi
    elif [[ "${release}" == "rocky" ]]; then
        if [[ ${os_version} -lt 95 ]]; then
            echo -e "${red} 请使用Rocky Linux 9.5或以上版本!${plain}\n" && exit 1
        fi
    elif [[ "${release}" == "ol" ]]; then
        if [[ ${os_version} -lt 8 ]]; then
            echo -e "${red} 请使用Oracle Linux 8或以上版本!${plain}\n" && exit 1
        fi
    else
        echo -e "${red}您的操作系统不支持此脚本.${plain}\n"
        echo "请确保您正在使用以下受支持的操作系统之一:"
        echo "- Ubuntu 22.04+"
        echo "- Debian 12+"
        echo "- CentOS 9+"
        echo "- Fedora 36+"
        echo "- Arch Linux"
        echo "- Parch Linux"
        echo "- Manjaro"
        echo "- Armbian"
        echo "- AlmaLinux 9.5+"
        echo "- Rocky Linux 9.5+"
        echo "- Oracle Linux 8+"
        echo "- OpenSUSE Tumbleweed"
        exit 1
    fi
}

check_pmc(){
    check_release
    if [[ "$release" == "debian" || "$release" == "ubuntu" || "$release" == "kali" ]]; then
        updates="apt update -y"
        installs="apt install -y"
        apps=("wget" "curl" "tar")
    elif [[ "$release" == "opensuse-tumbleweed" ]]; then
        updates="zypper refresh"
        installs="zypper install -y"
        apps=("wget" "curl" "tar")
    elif [[ "$release" == "almalinux" || "$release" == "centos" || "$release" == "rocky" || "$release" == "oracle" ]]; then
        updates="dnf update -y"
        installs="dnf install -y"
        apps=("wget" "curl" "tar")
    elif [[ "$release" == "fedora" ]]; then
        updates="dnf update -y"
        installs="dnf install -y"
        apps=("wget" "curl" "tar")
    elif [[ "$release" == "arch" || "$release" == "manjaro" || "$release" == "parch"  ]]; then
        updates="pacman -Syu"
        installs="pacman -Syu --noconfirm"
        apps=("wget" "curl" "tar")
    fi
}

install_base(){
    check_pmc
    cmds=("wget" "curl" "tar")
    echo -e "${green}[Info]${plain} 你的系统是${red} $release $os_version ${plain}"
    echo

    for i in "${!cmds[@]}"; do
        if ! which "${cmds[i]}" &>/dev/null; then
            DEPS+=("${apps[i]}")
        fi
    done
    
    if [ ${#DEPS[@]} -gt 0 ]; then
        echo -e "${yellow}[Tip]${plain} 安装依赖列表：${green}${DEPS[*]}${plain} 请稍后..."
        $updates 
        $installs "${DEPS[@]}" 
    else
        echo -e "${yellow}[Tip]${plain} 所有依赖已存在，不需要额外安装。"
    fi
}

config_after_install(){
    echo -e "${yellow}Migration... ${plain}"
    /usr/local/s-ui/sui migrate &>/dev/null
    
    echo -e "${yellow}安装/更新完成！出于安全考虑，建议修改面板设置。 ${plain}"
    read -p "您是否要继续进行修改 [y/n]？ ": config_confirm
    if [[ "${config_confirm}" == "y" || "${config_confirm}" == "Y" ]]; then
        echo -e "请输入${yellow}面板端口${plain} (默认值则留空）:"
        read config_port
        echo -e "请输入${yellow}面板路径${plain} (默认值则留空):"
        read config_path

        # Sub configuration
        echo -e "请输入${yellow}订阅端口${plain} (默认值则留空):"
        read config_subPort
        echo -e "请输入${yellow}订阅路径${plain} (默认值则留空):" 
        read config_subPath

        # Set configs
        echo -e "${yellow}正在初始化，请稍候...${plain}"
        params=""
        [ -z "$config_port" ] || params="$params -port $config_port"
        [ -z "$config_path" ] || params="$params -path $config_path"
        [ -z "$config_subPort" ] || params="$params -subPort $config_subPort"
        [ -z "$config_subPath" ] || params="$params -subPath $config_subPath"
        /usr/local/s-ui/sui setting ${params}

        read -p "您是否要更改管理员凭据 [y/n]? ": admin_confirm
        if [[ "${admin_confirm}" == "y" || "${admin_confirm}" == "Y" ]]; then
            # First admin credentials
            read -p "请设置您的用户名:" config_account
            read -p "请设置您的密码:" config_password

            # Set credentials
            echo -e "${yellow}正在初始化，请稍候...${plain}"
            /usr/local/s-ui/sui admin -username ${config_account} -password ${config_password}
        else
            echo -e "${yellow}您当前的管理员凭据: ${plain}"
            /usr/local/s-ui/sui admin -show
        fi
    else
        echo -e "${red}cancel...${plain}"
        if [[ ! -f "/usr/local/s-ui/db/s-ui.db" ]]; then
            local usernameTemp=$(head -c 6 /dev/urandom | base64)
            local passwordTemp=$(head -c 6 /dev/urandom | base64)
            echo -e "这是全新安装，出于安全考虑将生成随机登录信息:"
            echo -e "###############################################"
            echo -e "${green}用户名:${usernameTemp}${plain}"
            echo -e "${green}密码:${passwordTemp}${plain}"
            echo -e "###############################################"
            echo -e "${red}如果您忘记了登录信息,您可以输入 ${green}s-ui${red} 进入配置菜单${plain}"
            /usr/local/s-ui/sui admin -username ${usernameTemp} -password ${passwordTemp}
        else
            echo -e "${red} 这是您的升级，将保留原有设置,如果您忘记了登录信息,您可以输入 ${green}s-ui${red} 进入配置菜单${plain}"
        fi
    fi
}

prepare_services(){
    if [[ -f "/etc/systemd/system/sing-box.service" ]]; then
        echo -e "${yellow}停止 sing-box 服务... ${plain}"
        systemctl stop sing-box
        rm -f /usr/local/s-ui/bin/sing-box /usr/local/s-ui/bin/runSingbox.sh /usr/local/s-ui/bin/signal
    fi
    if [[ -e "/usr/local/s-ui/bin" ]]; then
        echo -e "###############################################################"
        echo -e "${green}/usr/local/s-ui/bin${red} 目录是否存在!"
        echo -e "请检查内容并在迁移后手动删除。 ${plain}"
        echo -e "###############################################################"
    fi
    systemctl daemon-reload
}

install_s-ui(){
    check_arch
    cd /tmp/

    if [ $# == 0 ]; then
        last_version=$(curl -Ls "https://api.github.com/repos/elesssss/s-ui/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [[ ! -n "$last_version" ]]; then
            echo -e "${red}无法获取 s-ui 版本，这可能是由于 GitHub API 限制所致，请稍后再试${plain}"
            exit 1
        fi
        echo -e "已获取S-UI最新版本: ${last_version}, 开始安装..."
        wget -N --no-check-certificate -O /tmp/s-ui-linux-${arch}.tar.gz https://github.com/elesssss/s-ui/releases/download/${last_version}/s-ui-linux-${arch}.tar.gz
        if [[ $? -ne 0 ]]; then
            echo -e "${red}下载 s-ui 失败, 请确保您的服务器能够访问GitHub ${plain}"
            exit 1
        fi
    else
        last_version=$1
        url="https://github.com/elesssss/s-ui/releases/download/${last_version}/s-ui-linux-${arch}.tar.gz"
        echo -e "开始安装 s-ui v$1"
        wget -N --no-check-certificate -O /tmp/s-ui-linux-${arch}.tar.gz ${url}
        if [[ $? -ne 0 ]]; then
            echo -e "${red}下载 s-ui v$1 失败，请确认该版本是否存在。${plain}"
            exit 1
        fi
    fi

    if [[ -e /usr/local/s-ui/ ]]; then
        systemctl stop s-ui
    fi

    tar zxvf s-ui-linux-${arch}.tar.gz
    rm s-ui-linux-${arch}.tar.gz -f

    wget -O /usr/bin/s-ui -N --no-check-certificate https://raw.githubusercontent.com/elesssss/s-ui/main/s-ui.sh
    chmod +x /usr/bin/s-ui
    cp -rf s-ui /usr/local/
    cp -f s-ui/*.service /etc/systemd/system/
    rm -rf s-ui

    config_after_install
    prepare_services

    systemctl enable s-ui --now

    echo -e "${green}s-ui v${last_version}${plain} 安装完成，现在已经正常运行了..."
    echo -e "您可以通过以下URL访问该面板(s):${green}"
    /usr/local/s-ui/sui uri
    echo -e "${plain}"
    echo -e ""
    s-ui help
}
echo -e "${green}正在执行...${plain}"
install_base
install_s-ui $1
