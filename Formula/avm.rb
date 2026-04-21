class Avm < Formula
  desc "Lightweight local/global command alias manager (like asdf/nvm for aliases)"
  homepage "https://github.com/DracoRunner/avm"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DracoRunner/avm/releases/download/v0.1.0/avm_darwin_arm64.tar.gz"
      sha256 "9cb4b56a2e5534e364fd2e870d8b8ac70a175378005936b5bbb65d07171e2af6"
    else
      url "https://github.com/DracoRunner/avm/releases/download/v0.1.0/avm_darwin_amd64.tar.gz"
      sha256 "d7c95fc37f920492eb5aa27cd7543b479d102bc156c09d9fdde59124ec2aa4c7"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DracoRunner/avm/releases/download/v0.1.0/avm_linux_arm64.tar.gz"
      sha256 "febfe3d29db3c127b0af15fba7a9e121a04a53391a41a9017fdc2fa2e1c48caf"
    else
      url "https://github.com/DracoRunner/avm/releases/download/v0.1.0/avm_linux_amd64.tar.gz"
      sha256 "3b552a3750a5eb24ffa6f839875831652d57e9ad6675ff118671e99010c9b08f"
    end
  end

  def install
    bin.install "avm" => "avm-bin"
  end

  def caveats
    <<~EOS
      To enable avm in your shell, add this to your ~/.zshrc or ~/.bashrc:

        eval "$(avm-bin shell-init)"

      Then reload your shell:

        source ~/.zshrc  # or source ~/.bashrc
    EOS
  end

  test do
    system "#{bin}/avm-bin", "version"
  end
end
