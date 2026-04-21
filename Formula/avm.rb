class Avm < Formula
  desc "A lightweight local/global command alias manager"
  homepage "https://github.com/DracoRunner/avm"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DracoRunner/avm/releases/download/v0.1.0/avm_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    else
      url "https://github.com/DracoRunner/avm/releases/download/v0.1.0/avm_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/DracoRunner/avm/releases/download/v0.1.0/avm_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    else
      url "https://github.com/DracoRunner/avm/releases/download/v0.1.0/avm_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
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
