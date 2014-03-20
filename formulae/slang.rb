require 'formula'

class Slang < Formula
  
  env :std # this is necessary for CGO, apparently
  
  url 'https://github.com/bww/Slang.git'
  homepage 'https://github.com/bww/Slang'
  version '2'
  
  depends_on 'go'
  
  def install
    # build Slang
    system "make"
    # install Slang
    bin.install("slang")
    # install resources here...
    share.install("resources/slang")
  end
  
end

