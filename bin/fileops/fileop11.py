import os, sys, shutil, time, random
from random import randint
import hashlib
import itertools
import threading
from threading import Thread
from collections import OrderedDict

# FileOps class contain the various methods for various file operations
class FileOps():
    def __init__(self):  
        self.Dirs = []  # this will store all directory names after creation
        self.NewDirs = [] # this will store directory names after rename 
        self.NestedDirs = [] # this will created nested directory under a TopLevelDir 
        self.NewNestedDirs = [] # this will rename the nested dirs
        self.SubDirs = []  # this will create subdirs inside a Dir 
        self.NewSubDirs = [] # this will rename subdirs 
        self.Files = []  # this will store all files inside a dir provided
        self.NewFiles = [] # this will store all files after renames  
        self.FilesAfterDirRename = [] # this will store all files after dir renames
        self.FilesCreatedBeforeDirRename = [] # this will store all files after dir renames
        self.FilesInSubDir = []    # this will store all files in subdir 
        self.FilesInNestedDir = [] # this will store all files in nested dir
        self.FilesInNewNestedDir = [] # this will store all files after nested dir rename 
        self.Md5Csum = {}  # dict with filename as the key to hold md5 checksum after file creation
        self.cc_drive1 = sys.argv[1]
        self.cc_drive2 = sys.argv[2]
        self.cc_drive3 = sys.argv[3]
        self.TopLevelDir = "Dir_" + time.strftime("%H%M%S")

    # method to create number_dirs dirs and save this in a list
    def create_dir(self,drive,number_dirs):    
        for i in range(number_dirs):        
            name = "Dir_" + time.strftime("%H%M%S") + "-" #appending timestamp to "Dir_"
            dir_name=name+str(i)
            self.Dirs.append(dir_name)
        for dir in self.Dirs:
            dir_full_path=os.path.join(drive, dir)
            if os.path.isdir(dir_full_path):
               print "Error: dir_full_path exists"
            else:
               os.mkdir(dir_full_path)
           
    def create_nested_dirs(self,drive,levels):
        tmp =[[self.TopLevelDir]]    # temp list storing the TopLevel dir
        #tmp =[]    # temp list storing the TopLevel dir
        dir_nested_path=[]
        for i in range(levels):
            tmp.append(["S_"+str(i)])           
        for item in itertools.product(*tmp):
             dir_full_path=os.path.join(drive,*item)
             #dir_nested_path=os.path.join(self.TopLevelDir,*item)
             dir_nested_path=os.path.join(*item)
             #print "Deb22" +dir_full_path
             #print "Deb33" +dir_nested_path
             tmp.append(dir_full_path)
             if os.path.isdir(dir_full_path):
                print "Error: {} exists".format(dir_full_path)
             else:
                os.makedirs(dir_full_path)
                # save the dirpath without the drive label
                #tmp_path =   dir_full_path[3:]
                #tmp_path = dir_nested_path
                #print "Deb10" +tmp_path 
         # also keep track of top level Dir
        self.Dirs.append(self.TopLevelDir)
        self.NestedDirs.append(dir_nested_path)

    def create_sub_dirs(self,cc_drive,dir):
        # Subdir will be created inside the dir 
        name = "SubDir_" + time.strftime("%H%M%S")
        top_dir_path=os.path.join(cc_drive,dir)
        dir_full_path=os.path.join(top_dir_path,name)
        if os.path.isdir(dir_full_path):
           print "Error: {} exists".format(dir_full_path)
        else:
           os.makedirs(dir_full_path)
           # save the dirpath without the drive label
           tmp_path =   dir_full_path[3:]
           self.SubDirs.append(tmp_path)
		   
    def rename_dir(self,drive):
        # "_new" will be appended to new name
        name = "_new"
        for dir in self.Dirs:
            dir_full_path=os.path.join(drive, dir)
            new_dir_full_path = dir_full_path+name
            if os.path.isdir(dir_full_path):
               #try:
                  os.rename(dir_full_path,new_dir_full_path)
                  print new_dir_full_path
                  # save the dirpath without the drive label
                  tmp_path = dir + name
                  #tmp_path =   new_dir_full_path[3:]
                  self.NewDirs.append(tmp_path)
                  # if the files are already created
                  tmp_file_path=[]
                  for dirname, dirnames, filenames in os.walk(new_dir_full_path):
                      tmp_file_path=filenames
                      print tmp_file_path
                  for file in tmp_file_path:
                      tmp_path = os.path.join(drive,dirname)
                      new_path = os.path.join(tmp_path,file)
                      print new_path
                      self.FilesCreatedBeforeDirRename.append(new_path)
               #except WindowsError:
               #   print "Permission or AccessDenied error reported, expected in rename in a multi CC setup"
            else:
               print "Error: " + dir_full_path + " does not exist"

    def rename_nested_dirs(self,drive):
        # "_new" will be appended to new name
        name = "_new"
        for dir in self.NestedDirs:
            dir_full_path=os.path.join(drive, dir)
            new_dir_full_path = dir_full_path+name
            if os.path.isdir(dir_full_path):
               os.rename(dir_full_path,new_dir_full_path)    
               # save the dirpath without the drive label
               tmp_path = dir + name
               #print "Deb66" +tmp_path
               #tmp_path =   new_dir_full_path[3:]
               self.NewNestedDirs.append(tmp_path)  
            else:
               print "Error: " + dir_full_path + " does not exist"

    def rename_subdir(self,drive):
        # "_new" will be appended to new name
        name = "_new"
        for dir in self.SubDirs:
            dir_full_path=os.path.join(drive, dir)
            new_dir_full_path = dir_full_path+name
            if os.path.isdir(dir_full_path):
               os.rename(dir_full_path,new_dir_full_path)    
               # save the dirpath without the drive label
               tmp_path = dir + name
               #tmp_path =   new_dir_full_path[3:]
               self.NewSubDirs.append(tmp_path)  
            else:
               print "Error: " + dir_full_path + " does not exist"
		   
    def list_dir(self,drive, Dirs, number_dirs):
        count = 0
        for dir in Dirs:
            dir_full_path=os.path.join(drive, dir)
            if os.path.isdir(dir_full_path):
               for dirname, dirnames, filenames in os.walk(dir_full_path):
                   print dirname
                   count = count + 1            
            else:
               print "Error: " + dir_full_path + " does not exist"
        if count == number_dirs:
           print "PASS: All the directories created exist"          
        else:
           print "FAIL: All the directories created dont exist"          

    def remove_dir(self,drive,Dirs):  
        for dir in Dirs:
            dir_full_path=os.path.join(drive, dir)
            if os.path.isdir(dir_full_path):
               #os.rmdir(dir_full_path)
               shutil.rmtree(dir_full_path, ignore_errors=True)
            else:
               print "Error: " + dir_full_path + " does not exist"

    # returns a byte array of random bytes
    def randomBytes(self,n):
        return bytearray(random.getrandbits(8) for i in range(n))

    # returns the md5 checksum of the opened file
    def md5(self,fname):
        hash_md5 = hashlib.md5()
        with open(fname, "rb") as f:
             for chunk in iter(lambda: f.read(4096), b""):
                 hash_md5.update(chunk)            
        return hash_md5.hexdigest()


    # this will create file_names with .type extension. optional argument *threaded is for creating 
    # names with timestamps if multiple threads are used to create the files.
    def create_filenames(self,drive,dir,type,number_files,*threaded):
        for i in range(number_files):
            new_dir_full_path=os.path.join(drive, dir)
            if threaded:
               name = "file" + "-" + str(threaded[0]) + "-"            
            else:
               name = "file" + "-"
            file_name=name+str(i)+type
            file_path=os.path.join(new_dir_full_path, file_name)
            # save the file_path without the drive label
            #tmp_path =   file_path[3:]
            tmp_path = dir + "/" + file_name
            self.Files.append(tmp_path)          
            #print "Deb1 " +tmp_path
            #if there is a dir rename need to save in this list FilesAfterDirRename
            for dir in self.NewDirs:
                new_dir_full_path=os.path.join(drive, dir)
                name = "file" + "-"
                file_name=name+str(i)+type
                file_path=os.path.join(new_dir_full_path, file_name)			
                # save the file_path without the drive label
                #tmp_path1 =   file_path[3:]
                tmp_path1 = dir + "/" + file_name
                self.FilesAfterDirRename.append(tmp_path1)
            #if Subdir exists
            for dir in self.SubDirs:
                new_dir_full_path=os.path.join(drive, dir)
                name = "file" + "-"
                file_name=name+str(i)+type
                file_path=os.path.join(new_dir_full_path, file_name)			
                # save the file_path without the drive label
                #tmp_path1 =   file_path[3:]
                tmp_path1 = dir + "/" + file_name
                self.FilesInSubDir.append(tmp_path1)
            #if NestedDir exists
            for dir in self.NestedDirs:
                new_dir_full_path=os.path.join(drive, dir)
                name = "file" + "-"
                file_name=name+str(i)+type
                file_path=os.path.join(new_dir_full_path, file_name)			
                # save the file_path without the drive label
                #tmp_path1 =   file_path[3:]
                tmp_path1 = dir + "/" + file_name
                self.FilesInNestedDir.append(tmp_path1)
            #if NewNestedDir exists
            for dir in self.NewNestedDirs:
                new_dir_full_path=os.path.join(drive, dir)
                name = "file" + "-"
                file_name=name+str(i)+type
                file_path=os.path.join(new_dir_full_path, file_name)			
                # save the file_path without the drive label
                #tmp_path1 =   file_path[3:]
                tmp_path1 = dir + "/" + file_name
                self.FilesInNewNestedDir.append(tmp_path1)
            
    def create_file_and_calculate_csm(self,drive,bytes,*threaded):
        for file in self.Files:
            try:
               if threaded:
                  if str(threaded[0]) in file:
                     file_full_path=os.path.join(drive, file)
                     fl = open(file_full_path,'w')                 
                     rand_bytes = self.randomBytes(bytes)
                     fl.write(str(rand_bytes))
                     fl.close()
                     md5checksum = self.md5(file_full_path)
                     self.Md5Csum [file_full_path] = md5checksum              
               else:
                  file_full_path=os.path.join(drive, file)
                  fl = open(file_full_path,'w')
                  rand_bytes = self.randomBytes(bytes)
                  fl.write(str(rand_bytes))
                  fl.close()
                  md5checksum = self.md5(file_full_path)
                  self.Md5Csum [file_full_path] = md5checksum              
                  print " Filename is " + file_full_path + " and md5_checksum is " + md5checksum
            except Exception as e:
                 print "Error creating file " +str(e)
                 sys.exit(0)
        for file in self.FilesAfterDirRename:
            try:
                 file_full_path=os.path.join(drive, file)
                 fl = open(file_full_path,'w')
                 rand_bytes = self.randomBytes(bytes)
                 fl.write(str(rand_bytes))
                 fl.close()
                 md5checksum = self.md5(file_full_path)
                 self.Md5Csum [file_full_path] = md5checksum              
                 print " Filename is " + file_full_path + " and md5_checksum is " + md5checksum
            except Exception as e:
                 print "Error creating file " +str(e)
                 sys.exit(0)

		   
    def create_large_size_file_names(self,drive,dir,type,number_files):
        for i in range(number_files):
            new_dir_full_path=os.path.join(drive, dir)
            name = "file" + "-"
            file_name=name+str(i)+type
            file_path=os.path.join(new_dir_full_path, file_name)
            # save the file_path without the drive label
            #tmp_path =   file_path[3:]
            tmp_path = dir + "/" + file_name
            self.Files.append(tmp_path)          
		#if there is a dir rename need to save in this list FilesAfterDirRename
            for dir in self.NewDirs:
                new_dir_full_path=os.path.join(drive, dir)
                name = "file" + "-"
                file_name=name+str(i)+type
                file_path=os.path.join(new_dir_full_path, file_name)			
                # save the file_path without the drive label
                tmp_path1 =   file_path[3:]
                self.FilesAfterDirRename.append(tmp_path1)

    def create_large_size_file(self,drive,bytes):			
        for file in self.Files:
            try:
               file_full_path=os.path.join(drive, file)
               with open(file_full_path, "wb") as out:
                    out.truncate(bytes)
               md5checksum = self.md5(file_full_path)
               self.Md5Csum [file_full_path] = md5checksum
               print " Filename is " + file_full_path + " and md5_checksum is " + md5checksum
            except Exception as e:
               print "Error creating file " +str(e)
               sys.exit(0)
 
    def modify_files(self,drive,file,bytes):
           #if there is a dir rename need to save in this list FilesAfterDirRename
           if self.NewDirs:
               for dir in self.NewDirs:
               #taking out file_name, it would just return base file_name 
                  file_name =	 os.path.basename(file)
                  file_path=os.path.join(dir, file_name)	
                  file_full_path=os.path.join(drive,file_path)
                  # save the file_path without the drive label
                  #tmp_path1 =   file_full_path[3:]
                  tmp_path1 = file_path
                  #print "Deb77" +tmp_path1
           elif self.NewNestedDirs:
               for dir in self.NewNestedDirs:
               #taking out file_name, it would just return base file_name 
                  file_name =	 os.path.basename(file)
                  file_path=os.path.join(dir, file_name)	
                  file_full_path=os.path.join(drive,file_path)
                  # save the file_path without the drive label
                  #tmp_path1 =   file_full_path[3:]
                  tmp_path1 = file_path
                  #print "Deb77" +tmp_path1
           else:
               file_full_path=os.path.join(drive, file)
           try:
               fl = open(file_full_path,'w')
               rand_bytes = self.randomBytes(bytes)
               fl.write(str(rand_bytes))
               fl.close()
           except Exception as e:
               print "Error writing to file {}".format(file) 
               print str(e)
               sys.exit(0)

    def add_attributes(self,drive,dir):
        try:
           dir_path=os.path.join(drive,dir)
           os.system("attrib +a " +dir_path)
           os.system("attrib +r " +dir_path)
           os.system("attrib +h " +dir_path)
           #os.system("attrib +s " +file_full_path)
        except Exception as e:
           print "Error setting attribute to dir {}".format(dir_path) 
           print str(e)
           sys.exit(0)
		
    def remove_attributes(self,drive,dir):
        try:
           dir_path=os.path.join(drive, dir)
           os.system("attrib -a " +dir_path)
           os.system("attrib -r " +dir_path)
           os.system("attrib -h " +dir_path)
           #os.system("attrib -s " +file_full_path)
        except Exception as e:
           print "Error removing attribute to dir {}".format(dir_path) 
           print str(e)
           sys.exit(0)
		
    def add_acls(self,drive,dir):
        try:
           dir_path=os.path.join(drive,dir)
           #os.system("icacls " + dir_path + " /grant Everyone:f")
           os.system("icacls " + dir_path + " /grant user63:(OI)(CI)F")        
        except Exception as e:
           print "Error setting acls to dir {}".format(dir_path)
           print str(e)
           sys.exit(0)

    def remove_acls(self,drive,dir):
        try:
           dir_path=os.path.join(drive,dir)
           #os.system("icacls " + dir_path + " /remove Everyone:g")        
           os.system("icacls " + dir_path + " /remove user63:g")        
        except Exception as e:
           print "Error setting acls to dir {}".format(dir_path) 
           print str(e)
           sys.exit(0)
		
    def rename_files(self,drive,file):
        try:
           file_full_path=os.path.join(drive, file)	
           file_name_parts = file.split(".") #split actual filename and extension       	
           new_full_path=os.path.join(drive, file_name_parts[0]) + "_new." + file_name_parts[1] #constructing the new name
           os.rename(file_full_path,new_full_path)
           # save the file_path without the drive label
           tmp_path =   new_full_path[3:]
           self.NewFiles.append(tmp_path)        
        except Exception as e:
           print "Error renaming the file {}".format(file)
           print str(e)
           sys.exit(0)

    def delete_files(self,drive,file):
        try:
           file_full_path=os.path.join(drive, file)
           os.remove(file_full_path)        
        except Exception as e:
           print "Error deleting the file " +file_full_path
           print str(e)
           sys.exit(0)

    def __del__(self):
        self.TopLevelDir = " "

# this test creates n directories from lesse1(CC2), rename those directories from lesse1(CC2)
# create 5 .dat files from lesse1(CC2) and calculate md5 checksum, modify these files from lessor(CC1)
# recalculate md5 checksum and make sure that md5 checksum differs from the original, then delete  

def test1(number_dirs):      
    print "Test1 Start"
    print "========================================================= "
    my_fileops = FileOps()
    Md5Csum_After_Modify = {}   # this dict is for storing md5 checksum after file modification   
    # from cc2 create dir in cc1 (basically from lessse to lessor)
    print " From CC2 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive2,number_dirs)
    
    #creating 5 .dat files per directory each of size 0 bytes from lesse1
    print "From CC2 create 50 .txt files of 0 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.Dirs:
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".txt",50)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive2,0)
    time.sleep(5)
    # write some data (5k) to the files previously created from lessor
    print "From CC1 write some data to the files previously created"
    for file in my_fileops.Files:
         my_fileops.modify_files(my_fileops.cc_drive1,file,5000) 
 
    #calculate csum again 
    print "Recalculate md5 checksum after file modification"
    for file in my_fileops.Files:
        file_full_path=os.path.join(my_fileops.cc_drive3, file)
        md5checksum = my_fileops.md5(file_full_path)
        Md5Csum_After_Modify [file_full_path] = md5checksum
        print " Filename is " + file_full_path + " and md5_checksum is " + md5checksum 

    #store in sorted dictionary by keys
    Sorted_Md5Csum = OrderedDict(my_fileops.Md5Csum)
    Sorted_Md5Csum_After_Modify = OrderedDict(Md5Csum_After_Modify)
  
    #print both the dictionaries
    print "Print md5 checksum from both the dictionaries"    
    for key, value in Sorted_Md5Csum.items():
        print key, value
    for key, value in Sorted_Md5Csum_After_Modify.items():
        print key, value

    # pass criteria would be the values of two dictionaries Md5Csum and  
    # Md5Csum_After_Modify should not match as the files were modified in between
    for x_values, y_values in zip(Sorted_Md5Csum.iteritems(), Sorted_Md5Csum_After_Modify.iteritems()):
        if x_values == y_values:
           print 'PASS', x_values, y_values
        else:
           print 'FILES', x_values, y_values   

    time.sleep(5)	
    print " From CC2 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive2)
    time.sleep(5)
	
    # repeat file modification from lesse2
    # write some data (5k) to the files previously created from lesse2   
    print "From CC3 write some data to the files previously created"
    for file in my_fileops.FilesAfterDirRename:
        my_fileops.modify_files(my_fileops.cc_drive3,file,5000) 
  	   		   
    time.sleep(5)
    print " From CC1 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive1,my_fileops.NewDirs)

    print "Test1 End"
    print "========================================================= "

# repeat test1 with xls files
def test2(number_dirs):      
    print "Test2 Start"
    print "========================================================= "
    my_fileops = FileOps()
    Md5Csum_After_Modify = {}   # this dict is for storing md5 checksum after file modification   
    # from cc2 create dir in cc1 (basically from lessse to lessor)
    # from cc1 create dir in cc1
    print " From CC2 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive2,number_dirs)
    print " From CC2 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive2)

    #for dir in my_fileops.NewDirs:
    #    print dir

    #creating 5 .xls files per directory each of size 0 bytes from lesse1
    print "From CC2 create 5 .xls files of 0 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.NewDirs:
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".xls",5)
    #for file in my_fileops.FilesAfterDirRename:
    #    print file
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive2,0)
    time.sleep(5)
    # write some data (5k) to the files previously created from lessor
    print "From CC1 write some data to the files previously created"
    for file in my_fileops.FilesAfterDirRename:
         my_fileops.modify_files(my_fileops.cc_drive1,file,5000) 
 
    #calculate csum again 
    print "Recalculate md5 checksum after file modification"
    for file in my_fileops.FilesAfterDirRename:
        file_full_path=os.path.join(my_fileops.cc_drive1, file)
        md5checksum = my_fileops.md5(file_full_path)
        Md5Csum_After_Modify [file_full_path] = md5checksum
        print " Filename is " + file_full_path + " and md5_checksum is " + md5checksum 

    #store in sorted dictionary by keys
    Sorted_Md5Csum = OrderedDict(my_fileops.Md5Csum)
    Sorted_Md5Csum_After_Modify = OrderedDict(Md5Csum_After_Modify)
  
    #print both the dictionaries
    print "Print md5 checksum from both the dictionaries"    
    for key, value in Sorted_Md5Csum.items():
        print key, value
    for key, value in Sorted_Md5Csum_After_Modify.items():
        print key, value

    # pass criteria would be the values of two dictionaries Md5Csum and  
    # Md5Csum_After_Modify should not match as the files were modified in between
    for x_values, y_values in zip(Sorted_Md5Csum.iteritems(), Sorted_Md5Csum_After_Modify.iteritems()):
        if x_values == y_values:
           print 'FAIL', x_values, y_values
        else:
           print 'PASS', x_values, y_values   

    time.sleep(5)
    print " From CC2 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.NewDirs)
    print "Test2 End"
    print "========================================================= "

# CC-3187: Create_1000-Dir_CC2
# CC-3193:Delete_1000-Dir_CC2
def test3(number_dirs):    
    print "Test3 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC2 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive2,number_dirs)
    my_fileops.list_dir(my_fileops.cc_drive2,my_fileops.Dirs,number_dirs) 
    time.sleep(5)
    print "From CC3 create 1 .txt file of 1000 bytes in the previously created directory and calculate md5 checksum"
    for dir in my_fileops.Dirs:
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".txt",1)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive2,1000) 
    time.sleep(5)
    print "From CC3 rename the files"
    for file in my_fileops.Files:
         my_fileops.rename_files(my_fileops.cc_drive3,file)
    print " From CC3 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive3)
    print " From CC3 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive3,my_fileops.NewDirs)
    print "Test3 End"
    print "========================================================= "

# CC-3188:Create_1000-Dir_CC1
def test4(number_dirs):    
    print "Test4 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC2 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive2,number_dirs)
    time.sleep(5)
    print " From CC1 list those directories"    
    my_fileops.list_dir(my_fileops.cc_drive1,my_fileops.Dirs,number_dirs)
    time.sleep(5)
    print " From CC1 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive1)
    time.sleep(5)
    print " From CC2 list those directories" 
    my_fileops.list_dir(my_fileops.cc_drive2,my_fileops.NewDirs,number_dirs) 
    print " From CC3 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive3,my_fileops.NewDirs)
    print "Test4 End"
    print "========================================================= "


# CC-3192:Delete_1000_Files_CC2_Owner_CC1
# CC-3232:Dir_File_Operations_with_same_set_of_Files
def test5(number_files):
    print "Test5 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC1 create 1 directory"
    my_fileops.create_dir(my_fileops.cc_drive1,1)
    print " From CC1 list those directories"    
    my_fileops.list_dir(my_fileops.cc_drive1,my_fileops.Dirs,1)    
    print "From CC1 create number_files .dat files of 10000 bytes in the previously created directory and calculate md5 checksum"
    for dir in my_fileops.Dirs:
        my_fileops.create_filenames(my_fileops.cc_drive1,dir,".dat",number_files)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive1,10000)
    time.sleep(70)
    print " From CC2 delete the files"
    for file in my_fileops.Files:
         my_fileops.delete_files(my_fileops.cc_drive2,file)
    time.sleep(80)
    print "From CC1 re-create same set of files"
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive1,10)
    time.sleep(70)
    print " From CC2 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.Dirs)
    print "Test5 End"
    print "========================================================= "


def helper_create_files(self,drive,bytes,threaded):
    self.create_file_and_calculate_csm(drive,bytes,threaded)

# CC-3193:Delete_1000-Dir_CC2
# multithreaded: this test creates number_dirs dirs and then using 3 different threads create files in those directories for 3 different CCs
def test6(number_dirs):
    print "Test6 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC1 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive1,number_dirs)    
    time.sleep(2)
    #creating 2 .txt files per directory each of size 10 bytes from lessor, lesse1, lesse2
    print "From 3 CCs create 2 .dat files of 100 bytes in each of the previously created directories and calculate md5 checksum"
    # create the file names with timestmap and rand int
    tstamp1 = time.strftime("%H%M%S") + str(randint(0,100))
    time.sleep(1) 
    tstamp2 = time.strftime("%H%M%S") + str(randint(0,100))
    time.sleep(1)
    tstamp3 = time.strftime("%H%M%S") + str(randint(0,100))

    for dir in my_fileops.Dirs:
        my_fileops.create_filenames(my_fileops.cc_drive1,dir,".txt",2,tstamp1)
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".txt",2,tstamp2)
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".txt",2,tstamp3)
    t1= threading.Thread(target=helper_create_files, args=(my_fileops,my_fileops.cc_drive1,5000,tstamp1))
    t2= threading.Thread(target=helper_create_files, args=(my_fileops,my_fileops.cc_drive2,5000,tstamp2))
    t3= threading.Thread(target=helper_create_files, args=(my_fileops,my_fileops.cc_drive3,5000,tstamp3))    
    # start 3 threads
    t1.start()
    t2.start()
    t3.start()
    #wait for the threads to finish
    t1.join()
    t2.join()
    t3.join()
    time.sleep(2)
    print "From CC1 delete the files"
    for file in my_fileops.Files:
        my_fileops.delete_files(my_fileops.cc_drive1,file)
    time.sleep(2)
    print "From 3 CCs re-create same set of files"
    t4= threading.Thread(target=helper_create_files, args=(my_fileops,my_fileops.cc_drive1,5000,tstamp1))
    t5= threading.Thread(target=helper_create_files, args=(my_fileops,my_fileops.cc_drive2,5000,tstamp2))
    t6= threading.Thread(target=helper_create_files, args=(my_fileops,my_fileops.cc_drive3,5000,tstamp3))
    # start 3 threads
    t4.start()
    t5.start()
    t6.start()
    #wait for the threads to finish
    t4.join()
    t5.join()
    t6.join()
    time.sleep(4)	
    print " From CC2 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.Dirs)
    print "Test6 End"
# multithreaded 
def test61(number_dirs):
    print "Test61 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC1 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive1,number_dirs)    
    time.sleep(70)
    #creating 2 .dat files per directory each of size 10 bytes from lessor, lesse1, lesse2
    print "From 3 CCs create 2 .dat files of 100 bytes in each of the previously created directories and calculate md5 checksum"
    tstamp1 = time.strftime("%H%M%S") + str(randint(0,100))
    tstamp11 = time.strftime("%H%M%S") + str(randint(100,200))	
    time.sleep(1) 
    tstamp2 = time.strftime("%H%M%S") + str(randint(0,100))
    tstamp22 = time.strftime("%H%M%S") + str(randint(100,200))
    time.sleep(1)
    tstamp3 = time.strftime("%H%M%S") + str(randint(0,100))
    tstamp33 = time.strftime("%H%M%S") + str(randint(100,200))

    for dir in my_fileops.Dirs:
        my_fileops.create_filenames(my_fileops.cc_drive1,dir,".dat",2,tstamp1)
        my_fileops.create_filenames(my_fileops.cc_drive1,dir,".dat",2,tstamp11)
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".dat",2,tstamp2)
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".dat",2,tstamp22)	
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".dat",2,tstamp3)
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".dat",2,tstamp33)
		
    t1= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive1,10,tstamp1))
    t11= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive1,10,tstamp11))
    t2= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive2,10,tstamp2))
    t22= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive2,10,tstamp22))
    t3= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive3,10,tstamp3))    
    t33= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive3,10,tstamp33))
    t1.start()
    t2.start()
    t3.start()
    t11.start()
    t22.start()
    t33.start()    
    #wait for the threads to finish
    t1.join()
    t2.join()
    t3.join()
    t11.join()
    t22.join()
    t33.join()
    time.sleep(60)
    print "From CC1 delete the files"
    for file in my_fileops.Files:
        my_fileops.delete_files(my_fileops.cc_drive1,file)
    time.sleep(70)
    print "From 3 CCs re-create same set of files"
    t4= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive1,10,tstamp1))
    t44= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive1,10,tstamp11))
    t5= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive2,10,tstamp2))
    t55= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive2,10,tstamp22))
    t6= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive3,10,tstamp3))
    t66= threading.Thread(target=helper_create_files, args=(my_fileops.cc_drive3,10,tstamp33))
    t4.start()
    t5.start()
    t6.start()
    t44.start()
    t55.start()
    t66.start()
    #wait for the threads to finish
    t4.join()
    t5.join()
    t6.join()
    t44.join()
    t55.join()
    t66.join()
    time.sleep(70)	
    print " From CC2 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.Dirs)
    print "Test6 End"
    print "========================================================= "
# CC-3208: Claim_LargeFiles_of_size_10GB_CC1 (Large file)
# CC-3222:Create_Rename_Delete_Large_Files_CC2
# CC-3211:Reboot_CC1_During_Large_File_Claim (reboot when the test is run)
def test7(number_dirs,bytes):
    print "Test7 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC2 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive2,number_dirs)    
    print " From CC2 list those directories"    
    my_fileops.list_dir(my_fileops.cc_drive2,my_fileops.Dirs,number_dirs)   
    #creating 1 .dat files per directory each of size 10 GB from lesse1
    print "From CC2 create 1 .dat file of size {} bytes in each of the previously created directory and calculate md5 checksum".format(bytes) 
    for dir in my_fileops.Dirs:        
        my_fileops.create_large_size_file_names(my_fileops.cc_drive2,dir,".dat",1)
    my_fileops.create_large_size_file(my_fileops.cc_drive2,bytes)
    time.sleep(5)
    print "From CC3 rename the files"
    for file in my_fileops.Files:
         my_fileops.rename_files(my_fileops.cc_drive3,file)
    # write some data (5k) to the files previously created from lessor
    print "From CC3 write some data to the files previously created"
    for file in my_fileops.NewFiles:
         my_fileops.modify_files(my_fileops.cc_drive3,file,5000) 
    print " From CC2 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.Dirs)
    print "Test7 End"
    print "========================================================= "

# CC-3223:Create_Rename_Delete_Deep_Dirs_CC2
def test8(levels):
    print "Test8 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print "From CC2 create directories upto level {}".format(levels)
    my_fileops.create_nested_dirs(my_fileops.cc_drive2,levels)   
    time.sleep(5)
    print "From CC3 rename the bottom level dir"
    my_fileops.rename_nested_dirs(my_fileops.cc_drive3)    
    print "Get top level directory from CC3"
    top_level = my_fileops.cc_drive3 + my_fileops.TopLevelDir
    print top_level
    my_fileops.Dirs.append(top_level)
    time.sleep(5)
    print "From CC2 delete all directories"  
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.Dirs)
    print "Test8 End"
    print "========================================================= "

# CC-3225: Modifying_DOSAttributes_CC2   
def test9(levels):
    print "Test9 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print "From CC2 create directories upto level {}".format(levels)
    my_fileops.create_nested_dirs(my_fileops.cc_drive2,levels)   
    time.sleep(5)    
    #creating 10K .dat files per directory each of size 10 bytes from lesseor, lesse1, lesse2
    print "From CC3 create 10K .dat files of 10 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.NestedDirs:
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".dat",10000)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive3,10)		
    time.sleep(5)
    #modify the attributes on the top level directory
    print "From CC3 add the attributes to the top level directory"
    for dir in my_fileops.Dirs:        
        my_fileops.add_attributes(my_fileops.cc_drive3,dir) 
    time.sleep(5)
    #modify the attributes on the top level directory
    print "From CC2 remove the attributes to the top level directory"
    for dir in my_fileops.Dirs:
         my_fileops.remove_attributes(my_fileops.cc_drive2,dir) 
    time.sleep(5)
    print "From CC2 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.NestedDirs)
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.Dirs)
    print "Test9 End"
    print "========================================================= "

 # CC-3224:Modifying_NTACLS_CC2 
 
def test10(levels):
    print "Test10 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print "From CC2 create directories upto level {}".format(levels)
    my_fileops.create_nested_dirs(my_fileops.cc_drive2,levels)   
    time.sleep(5)    
    #creating 10K .dat files per directory each of size 10 bytes from lesseor, lesse1, lesse2
    print "From CC3 create 10K .dat files of 10 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.NestedDirs:
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".dat",10000)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive3,10)		
    time.sleep(5)	
    #add acls to top level directory
    print "From CC3 add the attributes to the top level directory"
    for dir in my_fileops.NestedDirs:
        my_fileops.add_acls(my_fileops.cc_drive3,dir) 
    time.sleep(5)
    #modify the attributes to top level directory
    print "From CC2 remove the attributes from the top level directory"
    for dir in my_fileops.Dirs:
         my_fileops.remove_acls(my_fileops.cc_drive2,dir) 
    print "From CC2 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.NestedDirs)
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.Dirs)
    print "Test10 End"
    print "========================================================= "

def test11(number_dirs):      
    print "Test11 Start"
    print "========================================================= "
    my_fileops = FileOps()
    Md5Csum_After_Modify = {}   # this dict is for storing md5 checksum after file modification   
    # from cc1 create dir in cc1
    print " From CC1 create directory ",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive1,number_dirs)
	       
    time.sleep(5)
    print " From CC2 rename the dir"
    my_fileops.rename_dir(my_fileops.cc_drive2)
   
    time.sleep(5)
    #creating 5 .dat files per directory each of size 0 bytes from lessor
    print "From CC1 create 5 .dat files of 0 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.NewDirs:
        my_fileops.create_filenames(my_fileops.cc_drive1,dir,".dat",5)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive1,0)      
    time.sleep(5)
    #write some data (5k) to the files previously created from lesse2
    print "From CC3 write some data to the files previously created"
    for file in my_fileops.FilesAfterDirRename:
        my_fileops.modify_files(my_fileops.cc_drive3,file,5000)  #optional parameter needs to be set to 1
        print file
  
    #calculate csum again 
    print "Recalculate md5 checksum after file modification"
    for file in my_fileops.FilesAfterDirRename:
        file_full_path=os.path.join(my_fileops.cc_drive2, file)
        md5checksum = my_fileops.md5(file_full_path)
        Md5Csum_After_Modify [file_full_path] = md5checksum
        print " Filename is " + file_full_path + " and md5_checksum is " + md5checksum 

    #store in sorted dictionary by keys
    Sorted_Md5Csum = OrderedDict(my_fileops.Md5Csum)
    Sorted_Md5Csum_After_Modify = OrderedDict(Md5Csum_After_Modify)
  
    #print both the dictionaries
    print "Print md5 checksum from both the dictionaries"    
    for key, value in Sorted_Md5Csum.items():
        print key, value
    for key, value in Sorted_Md5Csum_After_Modify.items():
        print key, value

    # pass criteria would be the values of two dictionaries Md5Csum and  
    # Md5Csum_After_Modify should not match as the files were modified in between
    for x_values, y_values in zip(Sorted_Md5Csum.iteritems(), Sorted_Md5Csum_After_Modify.iteritems()):
        if x_values == y_values:
           print 'FAIL', x_values, y_values
        else:
           print 'PASS', x_values, y_values   
	time.sleep(5)
	
    for dir in my_fileops.NewDirs:
        print "From CC2 create sub directories inside {} ".format(dir)
        my_fileops.create_sub_dirs(my_fileops.cc_drive2,dir)
	time.sleep(5)
	
	#creating 5 .dat files per directory each of size 0 bytes from lesse2
    print "From CC3 create 5 .dat files of 100 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.SubDirs:
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".dat",5)
	my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive3,100)
	
	# repeat file modification from lesse2
    # write some data (5k) to the files previously created from lesse2   
    print "From CC3 write some data to the files previously created"
    for file in my_fileops.FilesAfterDirRename:
        my_fileops.modify_files(my_fileops.cc_drive3,file,5000) 
    time.sleep(5)
    print "From CC2 rename the files"
    for file in my_fileops.FilesAfterDirRename:
         my_fileops.rename_files(my_fileops.cc_drive2,file)
    
    print " From CC1 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive1,my_fileops.NewDirs)
    print "Test11 End"
    print "========================================================= "

# CC-3203:Create_Rename_Delete_Recreate_FIle-CC2
def test12(number_dirs):
    print "Test12 Start"
    print "========================================================= "
    my_fileops = FileOps()
    # from cc1 create dir in cc1
    print " From CC3 create directory ",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive3,number_dirs)
	       
    time.sleep(5)
    #creating 5 .dat files per directory each of size 0 bytes from lessor
    print "From CC2 create 100 .dat files of 100 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.Dirs:
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".dat",100)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive2,100)      
    time.sleep(5)
    #write some data (5k) to the files previously created from lesse2
    print "From CC3 write some data to the files previously created"
    for file in my_fileops.Files:
        my_fileops.modify_files(my_fileops.cc_drive3,file,5000)  
		
    print "From CC3 rename the files"
    for file in my_fileops.Files:
         my_fileops.rename_files(my_fileops.cc_drive3,file)
    time.sleep(5)
    print " From CC1 delete the files"
    for file in my_fileops.NewFiles:
         my_fileops.delete_files(my_fileops.cc_drive1,file)
    time.sleep(5)
    print "From CC2 re-create same set of files"
    for dir in my_fileops.Dirs:
	    my_fileops.create_filenames(my_fileops.cc_drive2,dir,".dat",100)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive2,100)
    time.sleep(5)
    print " From CC2 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.Dirs)
    print "Test12 End"
    print "========================================================= "
 
# CCC-3195:Rename_1000-Dir_CC1
# CC-3194:Delete_1000-Dir_CC1
def test13(number_dirs):    
    print "Test13 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC2 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive2,number_dirs)
    my_fileops.list_dir(my_fileops.cc_drive2,my_fileops.Dirs,number_dirs) 
    time.sleep(5)
    print "From CC3 create 1 .dat file of 10 bytes in the previously created directory and calculate md5 checksum"
    for dir in my_fileops.Dirs:
        #create_file_and_calculate_csm(cc_drive2,dir,".dat",1, 10)
        my_fileops.create_large_size_file_names(my_fileops.cc_drive3,dir,".dat",1)
    my_fileops.create_large_size_file(my_fileops.cc_drive3,100)
    time.sleep(5)
    print "From CC1 rename the files"
    for file in my_fileops.Files:
         my_fileops.rename_files(my_fileops.cc_drive1,file)
    time.sleep(5)
    print " From CC2 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive2)
    my_fileops.list_dir(my_fileops.cc_drive2,my_fileops.NewDirs,number_dirs) 
    time.sleep(5)
    print " From CC3 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive3,my_fileops.NewDirs)
    print "Test13 End"
    print "========================================================= "

# CC-3196:Rename_1000-Dir_CC2
def test14(number_dirs):    
    print "Test14 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC3 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive3,number_dirs)
    my_fileops.list_dir(my_fileops.cc_drive3,my_fileops.Dirs,number_dirs) 
    time.sleep(5)
    print "From CC2 create 1 .dat file of 10 bytes in the previously created directory and calculate md5 checksum"
    for dir in my_fileops.Dirs:
        #create_file_and_calculate_csm(cc_drive2,dir,".dat",1, 10)
        my_fileops.create_large_size_file_names(my_fileops.cc_drive2,dir,".dat",1)
    my_fileops.create_large_size_file(my_fileops.cc_drive2,100)
    time.sleep(5)
    print "From CC1 rename the files"
    for file in my_fileops.Files:
         my_fileops.rename_files(my_fileops.cc_drive1,file)
    time.sleep(5)
    print " From CC2 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive2)
    my_fileops.list_dir(my_fileops.cc_drive2,my_fileops.NewDirs,number_dirs) 
    time.sleep(5)
    print " From CC1 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive1,my_fileops.NewDirs)
    print "Test14 End"
    print "========================================================= "

# CC-3189:Rename_1000_Files_CC2_Owner_CC3
# CC-3191:Delete_1000_Files_CC2_Owner_CC3
def test15(number_dirs):    
    print "Test15 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC3 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive3,number_dirs)
    my_fileops.list_dir(my_fileops.cc_drive3,my_fileops.Dirs,number_dirs) 
    print "From CC3 create 1 .dat file of 10 bytes in the previously created directory and calculate md5 checksum"
    for dir in my_fileops.Dirs:
        my_fileops.create_large_size_file_names(my_fileops.cc_drive3,dir,".dat",1)
    my_fileops.create_large_size_file(my_fileops.cc_drive3,100)
    time.sleep(5)
    print "From CC2 rename the files"
    for file in my_fileops.Files:
         my_fileops.rename_files(my_fileops.cc_drive2,file)
    time.sleep(5)
    print " From CC2 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive2)
    my_fileops.list_dir(my_fileops.cc_drive2,my_fileops.NewDirs,number_dirs) 
    time.sleep(5)
    print " From CC1 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive1,my_fileops.NewDirs)
    print "Test15 End"
    print "========================================================= "

# CC-3190:Rename_1000_Files_CC2_Owner_CC1
# CC-3192:Delete_1000_Files_CC2_Owner_CC1
def test16(number_dirs):    
    print "Test16 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC1 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive1,number_dirs)
    my_fileops.list_dir(my_fileops.cc_drive1,my_fileops.Dirs,number_dirs) 
    print "From CC1 create 1 .dat file of 10 bytes in the previously created directory and calculate md5 checksum"
    for dir in my_fileops.Dirs:
        my_fileops.create_large_size_file_names(my_fileops.cc_drive1,dir,".dat",1)
    my_fileops.create_large_size_file(my_fileops.cc_drive1,100)
    time.sleep(5)
    print "From CC2 rename the files"
    for file in my_fileops.Files:
         my_fileops.rename_files(my_fileops.cc_drive2,file)
    time.sleep(5)
    print " From CC2 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive2)
    my_fileops.list_dir(my_fileops.cc_drive2,my_fileops.NewDirs,number_dirs) 
    time.sleep(5)
    print " From CC3 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive3,my_fileops.NewDirs)
    print "Test16 End"
    print "========================================================= "
	
# CC-3209:CC-3209-Claim_LargeFiles_of_size_10GB_CC2
def test18(number_dirs,bytes):
    print "Test18 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print " From CC1 create directories",  number_dirs    
    my_fileops.create_dir(my_fileops.cc_drive1,number_dirs)
    print " From CC1 list those directories"    
    my_fileops.list_dir(my_fileops.cc_drive1,my_fileops.Dirs,number_dirs) 
    #creating 1 .dat files per directory each of size 10 GB from lessor
    print "From CC1 create 1 .dat file of size {} bytes in each of the previously created directory and calculate md5 checksum".format(bytes) 
    for dir in my_fileops.Dirs:
        my_fileops.create_large_size_file_names(my_fileops.cc_drive1,dir,".dat",1)
    my_fileops.create_large_size_file(my_fileops.cc_drive1,100)
    time.sleep(5)
    print "From CC2 rename the files"
    for file in my_fileops.Files:
         my_fileops.rename_files(my_fileops.cc_drive2,file)
    # write some data (5k) to the files previously created from lessor
    print "From CC2 write some data to the files previously created"
    my_fileops.rename_dir(my_fileops.cc_drive2)
    my_fileops.list_dir(my_fileops.cc_drive2,my_fileops.NewDirs,number_dirs) 
    for file in my_fileops.NewFiles:
         my_fileops.modify_files(my_fileops.cc_drive2,file,5000) 
    print " From CC2 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.NewDirs)
    print "Test18 End"
    print "========================================================= "

# CC-3197: Create_Rename_Recreate_CC1
def test19(number_dirs):
    print "Test19 Start"
    print "========================================================= "
    my_fileops = FileOps()
    Md5Csum_After_Modify = {}   # this dict is for storing md5 checksum after file modification   
    print " From CC1 create directory" , number_dirs   
    my_fileops.create_dir(my_fileops.cc_drive1,number_dirs)    
    print "From CC1 create 5 .dat file of size 100 bytes in each of the previously created directory and calculate md5 checksum".format(bytes) 
    for dir in my_fileops.Dirs:        
        my_fileops.create_filenames(my_fileops.cc_drive1,dir,".dat",5)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive1,100)
    time.sleep(5)
    print " From CC2 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive2)	
    time.sleep(5)
	# repeat file modification from lesse2
    # write some data (5k) to the files previously created from lesse2   
    print "From CC3 write some data to the files previously created"
    for file in my_fileops.FilesCreatedBeforeDirRename:
        my_fileops.modify_files(my_fileops.cc_drive3,file,5000) 
    #calculate csum again 
    print "Recalculate md5 checksum after file modification"
    for file in my_fileops.FilesCreatedBeforeDirRename:
        file_full_path=os.path.join(my_fileops.cc_drive3, file)
        md5checksum = my_fileops.md5(file_full_path)
        Md5Csum_After_Modify [file_full_path] = md5checksum
        print " Filename is " + file_full_path + " and md5_checksum is " + md5checksum 

    #store in sorted dictionary by keys
    Sorted_Md5Csum = OrderedDict(my_fileops.Md5Csum)
    Sorted_Md5Csum_After_Modify = OrderedDict(Md5Csum_After_Modify)  
    #print both the dictionaries
    print "Print md5 checksum from both the dictionaries"    
    for key, value in Sorted_Md5Csum.items():
        print key, value
    for key, value in Sorted_Md5Csum_After_Modify.items():
        print key, value
    # pass criteria would be the values of two dictionaries Md5Csum and  
    # Md5Csum_After_Modify should not match as the files were modified in between
    for x_values, y_values in zip(Sorted_Md5Csum.iteritems(), Sorted_Md5Csum_After_Modify.iteritems()):
        if x_values == y_values:
           print 'FAIL', x_values, y_values
        else:
           print 'PASS', x_values, y_values        
		   
    time.sleep(5)	
    for dir in my_fileops.NewDirs:
        print "From CC2 create sub directories inside {} ".format(dir)
        my_fileops.create_sub_dirs(my_fileops.cc_drive2,dir)
	
    time.sleep(5)
    my_fileops.Files[:] = [] # need to empty it as the same list will be used for files creation inside subdirs
    #creating 5 .dat files per directory each of size 0 bytes from lesse1
    print "From CC2 create 5 .dat files of 100 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.SubDirs:        
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".dat",5)        
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive2,100)
    	
    print "From CC2 rename the files inside the subdirectory"
    for file in my_fileops.FilesInSubDir:
        my_fileops.rename_files(my_fileops.cc_drive2,file)
        
    time.sleep(5)
    print "From CC3 rename the subdir"
    my_fileops.rename_subdir(my_fileops.cc_drive3)	
    
    # need to empty it as the same list will be used for subdir creation
    my_fileops.SubDirs[:] = [] 	
    for dir in my_fileops.NewSubDirs:
        print "From CC3 create sub directories inside {} ".format(dir)
        my_fileops.create_sub_dirs(my_fileops.cc_drive3,dir)  
     
    my_fileops.Files[:] = [] # need to empty it as the same list will be used for files creation inside subdirs
    #creating 5 .dat files per directory each of size 0 bytes from lesse1
    print "From CC3 create 5 .dat files of 100 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.SubDirs:        
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".dat",5)        
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive3,100)	 
	
	# repeat file modification from lesse2
    # write some data (5k) to the files previously created from lesse2   
    print "From CC3 write some data to the files previously created"
    for file in my_fileops.Files:
        my_fileops.modify_files(my_fileops.cc_drive3,file,5000) 
         
    print "From CC3 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive2,my_fileops.NewDirs)
    print "Test19 End"
    print "========================================================= "

# CC-3198:Create_Rename_Recreate_CC2
def test20(number_dirs):
    print "Test20 Start"
    print "========================================================= "
    my_fileops = FileOps()
    Md5Csum_After_Modify = {}   # this dict is for storing md5 checksum after file modification   
    print " From CC2 create directory" , number_dirs   
    my_fileops.create_dir(my_fileops.cc_drive2,number_dirs)    
    print "From CC2 create 5 .dat file of size 100 bytes in each of the previously created directory and calculate md5 checksum".format(bytes) 
    for dir in my_fileops.Dirs:        
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".dat",5)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive2,100)
    time.sleep(5)
    print " From CC3 rename the dirs"
    my_fileops.rename_dir(my_fileops.cc_drive3)	
    time.sleep(5)
	# repeat file modification from lesse2
    # write some data (5k) to the files previously created from lesse2   
    print "From CC3 write some data to the files previously created"
    for file in my_fileops.FilesCreatedBeforeDirRename:
        my_fileops.modify_files(my_fileops.cc_drive3,file,5000)

    #calculate csum again 
    print "Recalculate md5 checksum after file modification"
    for file in my_fileops.FilesCreatedBeforeDirRename:
        file_full_path=os.path.join(my_fileops.cc_drive3, file)
        md5checksum = my_fileops.md5(file_full_path)
        Md5Csum_After_Modify [file_full_path] = md5checksum
        print " Filename is " + file_full_path + " and md5_checksum is " + md5checksum 

    #store in sorted dictionary by keys
    Sorted_Md5Csum = OrderedDict(my_fileops.Md5Csum)
    Sorted_Md5Csum_After_Modify = OrderedDict(Md5Csum_After_Modify)  
    #print both the dictionaries
    print "Print md5 checksum from both the dictionaries"    
    for key, value in Sorted_Md5Csum.items():
        print key, value
    for key, value in Sorted_Md5Csum_After_Modify.items():
        print key, value
    # pass criteria would be the values of two dictionaries Md5Csum and  
    # Md5Csum_After_Modify should not match as the files were modified in between
    for x_values, y_values in zip(Sorted_Md5Csum.iteritems(), Sorted_Md5Csum_After_Modify.iteritems()):
        if x_values == y_values:
           print 'FAIL', x_values, y_values
        else:
           print 'PASS', x_values, y_values        
	   
    time.sleep(5)	
    for dir in my_fileops.NewDirs:
        print "From CC3 create sub directories inside {} ".format(dir)
        my_fileops.create_sub_dirs(my_fileops.cc_drive3,dir)
	
    time.sleep(5)
    my_fileops.Files[:] = [] # need to empty it as the same list will be used for files creation inside subdirs
    #creating 5 .dat files per directory each of size 0 bytes from lesse1
    print "From CC3 create 5 .dat files of 100 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.SubDirs:        
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".dat",5)        
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive3,100)
	
    time.sleep(5)	
    print "From CC2 rename the files inside the subdirectory"
    for file in my_fileops.FilesInSubDir:
        my_fileops.rename_files(my_fileops.cc_drive2,file)
        
    time.sleep(5)
    print "From CC2 rename the subdir"
    my_fileops.rename_subdir(my_fileops.cc_drive2)	
    
    # need to empty it as the same list will be used for subdir creation
    my_fileops.SubDirs[:] = [] 	
    for dir in my_fileops.NewSubDirs:
        print "From CC2 create sub directories inside {} ".format(dir)
        my_fileops.create_sub_dirs(my_fileops.cc_drive2,dir)  
     
    my_fileops.Files[:] = [] # need to empty it as the same list will be used for files creation inside subdirs
    #creating 5 .dat files per directory each of size 0 bytes from lesse1
    print "From CC2 create 5 .dat files of 100 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.SubDirs:        
        my_fileops.create_filenames(my_fileops.cc_drive2,dir,".dat",5)        
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive2,100)	 
	
	# repeat file modification from lesse2
    # write some data (5k) to the files previously created from lesse2   
    print "From CC2 write some data to the files previously created"
    for file in my_fileops.Files:
        my_fileops.modify_files(my_fileops.cc_drive2,file,5000) 
         
    print "From CC1 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive1,my_fileops.NewDirs)
    print "Test20 End"
    print "========================================================= "

# nested directory tests 
def test21(levels):
    print "Test21 Start"
    print "========================================================= "
    my_fileops = FileOps()
    print "From CC2 create directories upto level {}".format(levels)
    my_fileops.create_nested_dirs(my_fileops.cc_drive2,levels)   
    time.sleep(5)    
    #creating 10K .dat files per directory each of size 10 bytes from lesseor, lesse1, lesse2
    print "From CC3 create 10000 .dat files of 100 bytes in each of the previously created directories and calculate md5 checksum"
    for dir in my_fileops.NestedDirs:
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".dat",10000)
    my_fileops.create_file_and_calculate_csm(my_fileops.cc_drive3,100)		
    time.sleep(5)	
    print "From CC3 rename the bottom level dir"
    my_fileops.rename_nested_dirs(my_fileops.cc_drive3)    
    time.sleep(5)
    print "Recreate the filenames"
    for dir in my_fileops.NewNestedDirs:
        my_fileops.create_filenames(my_fileops.cc_drive3,dir,".dat",10000)
    print "From CC1 write some data to the files previously created"
    for file in my_fileops.FilesInNewNestedDir:
         my_fileops.modify_files(my_fileops.cc_drive1,file,5) 
    time.sleep(5)
    print " From CC2 delete the files"
    for file in my_fileops.FilesInNewNestedDir:
         my_fileops.delete_files(my_fileops.cc_drive2,file)
    time.sleep(5)
    print "From CC3 delete all directories" 
    my_fileops.remove_dir(my_fileops.cc_drive3,my_fileops.Dirs)
    print "Test21 End"
    print "========================================================= "


def main():
     #test1(2)
     #test2(5)
     #test3(10000) #14cc setup issue
     #test4(1000)
     #test5(10000)
     test6(10000)
     #test61(1000)
     '''
     test7(2, 10 * 1024 *1024 * 1024) #10 GB
     test7(1,20 * 1024 *1024 * 1024) #20 GB
     test8(20)
     test9(10)
     test10(10)
     test11(1)
     test12(5)
     test13(1000)
     test14(1000)
     test15(1000)
     test16(1000) #bug 15682
     test19(1)
     test20(1)
     '''
     #test21(100)

main()
