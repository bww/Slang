require 'formula'

class Slang < Formula
  
  url 'https://github.com/bww/Slang.git'
  homepage 'https://github.com/bww/Slang'
  version '1'
  
  def install
    # build Slang
    system "make"
    # install Slang
    bin.install("slang")
  end
  
end

